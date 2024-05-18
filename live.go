package main

import (
	"agent/chrome"
	"agent/helper"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func collectCommandFiles(sessionPath string, completedCommands helper.Set[string]) []string {

	pth := filepath.Join(sessionPath, "commands")
	dirSlice, err := os.ReadDir(pth)

	if err != nil {
		log.Fatal(err)
	}

	type fileCommand struct {
		modTime  time.Time
		filename string
	}

	newCommands := make([]fileCommand, 0, 3)

	for _, el := range dirSlice {
		if !el.IsDir() {

			if !completedCommands.Has(filepath.Join(pth, el.Name())) && strings.HasSuffix(el.Name(), ".json") {

				info, err := el.Info()

				if err != nil {
					log.Fatal(err)
				}

				newCommands = append(newCommands, fileCommand{
					modTime:  info.ModTime(),
					filename: filepath.Join(pth, el.Name()),
				})
			}
		}
	}

	sort.Slice(newCommands, func(i, j int) bool {
		return newCommands[i].modTime.Unix() < newCommands[j].modTime.Unix()
	})

	retCommands := make([]string, len(newCommands))

	for index, com := range newCommands {
		retCommands[index] = com.filename
	}

	return retCommands
}

func performAction(
	ctx context.Context,
	action chromedp.Action,
	job *chrome.FileJob,
	commandDurationMs *uint64,
) error {

	newCtx := context.Background()
	if commandDurationMs != nil {
		var cancelFunc context.CancelFunc
		dur := time.Millisecond * time.Duration(*commandDurationMs)
		newCtx, cancelFunc = context.WithTimeout(newCtx, dur)
		defer cancelFunc()
	}

	errChan := make(chan error)

	go func() {
		err := action.Do(ctx)
		errChan <- err
	}()

	select {
	case resp := <-errChan:
		if resp != nil {
			fmt.Println(resp)
			return resp
		}
	case <-newCtx.Done():
		return errors.New("command context deadline exceeded")
	}

	go func() {
		job.GetWaitGroup().Wait()
		close(job.GetChannel())
	}()

	var err error

	for err = range job.GetChannel() {
	}
	return err
}

func writeErr(writePath string, err error) error {
	writeBytes := []byte(err.Error())
	writePath = filepath.Join(writePath, "err.txt")
	err = os.WriteFile(writePath, writeBytes, 0777)
	return err
}

func processOperations(
	filePath string,
	ctx context.Context,
	sessionPath string,
	waitTime *uint64) error {

	byteSlice, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	_, fname := filepath.Split(filePath)
	nameWoExt := strings.Split(fname, ".")[0]

	filePath = filepath.Join(sessionPath, "responses", nameWoExt)

	if err = os.Mkdir(filePath, 0777); err != nil {
		log.Fatal(err)
	}

	op := &Operation{}

	err = json.Unmarshal(byteSlice, op)
	if err != nil {
		fmt.Println(len(byteSlice))
		fmt.Println(string(byteSlice))
		fmt.Println(err)
		return err
	}

	job := chrome.InitFileJob()

	var responseErr error

	switch op.Type {
	case "browser":
		lastCommand := op.CommandList[len(op.CommandList)-1]
		action := chrome.AddOperation(lastCommand.Params, lastCommand.CommandName, filePath, job)
		responseErr = performAction(ctx, action, job, waitTime)
	}

	if responseErr != nil {
		err := writeErr(filePath, responseErr)

		if err != nil {
			return err
		}
	}

	return nil
}

func getLiveSession(
	sessionPath string,
	waitTime *uint64) chromedp.ActionFunc {
	return func(c context.Context) error {
		commandSet := helper.Set[string]{}
		alive := true

		go func() {
			<-c.Done()
			alive = false
		}()

		for alive {
			commandSlice := collectCommandFiles(sessionPath, commandSet)

			for _, commandFileName := range commandSlice {
				err := processOperations(commandFileName, c, sessionPath, waitTime)

				if err != nil {
					alive = false
				}

				commandSet.Insert(commandFileName)
			}
		}

		return nil
	}
}

func createLiveFolders(sessionPath string) {
	commandPath := filepath.Join(sessionPath, "commands")
	responsePath := filepath.Join(sessionPath, "responses")
	if err := os.MkdirAll(commandPath, 0777); !os.IsExist(err) && err != nil {
		log.Fatal("Could not create directory: " + commandPath)
	}

	if err := os.MkdirAll(responsePath, 0777); !os.IsExist(err) && err != nil {
		log.Fatal("Could not create directory: " + responsePath)
	}
}

func RunLive(timeout uint64, headless bool, commandRunTime *uint64, sessionPath string) {
	var ctx context.Context
	var cancel context.CancelFunc

	createLiveFolders(sessionPath)

	if headless {
		ctx, _ = chromedp.NewContext(
			context.Background(),
		)
	} else {
		actx, _ := chromedp.NewExecAllocator(
			context.Background(),
			append(
				chromedp.DefaultExecAllocatorOptions[:],
				chromedp.Flag("headless", false))...)

		ctx, _ = chromedp.NewContext(
			actx,
		)
	}

	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	tasks := chromedp.Tasks{
		getLiveSession(sessionPath, commandRunTime),
	}

	_ = chromedp.Run(ctx, tasks)
}
