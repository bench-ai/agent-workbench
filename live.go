package main

import (
	"agent/chrome"
	"agent/helper"
	"context"
	"encoding/json"
	"errors"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

/*
Here you will find functions used to live sessions, along with helper functions tp help indicate processes completed running
when making wrapper libraries
*/

/*
collectCommandFiles

collects all the commands that have been sent to the live session
*/
func collectCommandFiles(sessionPath string, completedCommands helper.Set[string]) ([]string, error) {

	pth := filepath.Join(sessionPath, "commands")
	dirSlice, err := os.ReadDir(pth)

	if err != nil {
		return nil, err
	}

	type fileCommand struct {
		modTime  time.Time
		filename string
	}

	newCommands := make([]fileCommand, 0, 3)

	/*
		in this for loop we will collect all unperformed live commands
	*/
	for _, el := range dirSlice {
		if !el.IsDir() {

			if !completedCommands.Has(filepath.Join(pth, el.Name())) && strings.HasSuffix(el.Name(), ".json") { // live commands will always be json files

				info, err := el.Info()

				if err != nil {
					return nil, err
				}

				newCommands = append(newCommands, fileCommand{
					modTime:  info.ModTime(),
					filename: filepath.Join(pth, el.Name()),
				})
			}
		}
	}

	/*
		we will sort the commands from earliest to latest
	*/
	sort.Slice(newCommands, func(i, j int) bool {
		return newCommands[i].modTime.Unix() < newCommands[j].modTime.Unix()
	})

	retCommands := make([]string, len(newCommands))

	for index, com := range newCommands {
		retCommands[index] = com.filename
	}

	return retCommands, nil
}

/*
performAction

browser commands usually write files, the way file writing is handled is the job struct schedules file
writing to be run by a background process allowing multiple files to be written in the background at once.

In a live context where choosing commands is dynamic we generally need the file data to make a decision on what command
to run next. These command forces the file data to finish writing after every command is executed
*/
func performAction(
	ctx context.Context,
	action chromedp.Action,
	job *chrome.FileJob,
	commandDurationMs *uint64,
) error {

	// initialize a context with a duration if necessary
	newCtx := context.Background()
	if commandDurationMs != nil {
		var cancelFunc context.CancelFunc
		dur := time.Millisecond * time.Duration(*commandDurationMs)
		newCtx, cancelFunc = context.WithTimeout(newCtx, dur)
		defer cancelFunc()
	}

	errChan := make(chan error)

	// run the command requested in the background and write the response to a channel upon completion
	go func() {
		err := action.Do(ctx)
		errChan <- err
	}()

	// wait for the command to finish or for the context to run out
	select {
	case resp := <-errChan:
		if resp != nil {
			return resp
		}
	case <-newCtx.Done():
		return errors.New("command context deadline exceeded")
	}

	// have a background process that waits for all file writing jobs to finish, and then closes the channel
	go func() {
		job.GetWaitGroup().Wait()
		close(job.GetChannel())
	}()

	var err error

	// runs until the job channel closes essentially stopping the function from exiting until all jobs are done

	for err = range job.GetChannel() {
	}

	return err
}

/*
writeErr

if a command fails some indication is needed that, that happened, this command writes an error file for that
purpose
*/
func writeErr(writePath string, err error) error {
	writeBytes := []byte(err.Error())
	writePath = filepath.Join(writePath, "err.txt")
	err = os.WriteFile(writePath, writeBytes, 0777)
	return err
}

/*
writeSuccess

indicates a command is done running
*/
func writeSuccess(writePath string) error {
	var writeBytes []byte
	writePath = filepath.Join(writePath, "success.txt")
	err := os.WriteFile(writePath, writeBytes, 0777)
	return err
}

/*
endSession

a text file that tells the agent to gracefully exit
*/
func endSession(sessionPath string, exitErr error) error {
	exitPath := filepath.Join(sessionPath, "exit.txt")
	err := os.WriteFile(exitPath, []byte(exitErr.Error()), 0777)
	return err
}

/*
processOperations

processes commands delivered to the agent
*/
func processOperations(
	filePath string,
	ctx context.Context,
	sessionPath string,
	waitTime *uint64) error {

	// read the command file
	byteSlice, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	_, fname := filepath.Split(filePath)
	nameWoExt := strings.Split(fname, ".")[0]
	filePath = filepath.Join(sessionPath, "responses", nameWoExt)

	// create a director for the command
	if err = os.Mkdir(filePath, 0777); err != nil {
		return err
	}

	op := &Operation{}

	err = json.Unmarshal(byteSlice, op)
	if err != nil {
		return err
	}

	job := chrome.InitFileJob()

	var responseErr error // error that signals the command failed

	/*
		response error will not be returned by this function, but any error that is, will automatically trigger the death
		of the session
	*/

	/**
	TODO: LLM operations, integrate tool calls, it should be able to extract info from the repsonse json
	add task for post processing loading
	*/

	switch op.Type {
	case "browser":
		lastCommand := op.CommandList[len(op.CommandList)-1]
		action, err := chrome.AddOperation(lastCommand.Params, lastCommand.CommandName, filePath, job)
		if err != nil {
			// if the action does not exist shut down the session
			return err
		}
		responseErr = performAction(ctx, action, job, waitTime)
	case "llm":
		//waitSeconds := *waitTime / 1000
		if *waitTime > 32767 {
			return errors.New("command wait time exceeds limit for LLM'S the limit is 32767 seconds")
		}

		llmWaitTime := int32(*waitTime)
		op.Settings.Timeout = &llmWaitTime
		responseErr = runLlmCommands(op.Settings, op.CommandList, filePath)
	case "exit":
		// an exit command was received so the session must end
		return errors.New("session has manually exited")
	}

	if responseErr != nil {
		err = writeErr(filePath, responseErr)
	} else {
		err = writeSuccess(filePath)
	}

	// if success or command errors cannot be written return that error

	if err != nil {
		return err
	}

	return nil
}

/*
getLiveSession

starts a live session in a chromedp action loop
*/
func getLiveSession(
	sessionPath string,
	waitTime *uint64) chromedp.ActionFunc {
	return func(c context.Context) error {
		commandSet := helper.Set[string]{}
		alive := true

		var exitErr error

		go func() {
			// this function runs as a background process, it waits for the context to exceed and
			//signals to the while loop that the session has ended
			<-c.Done()
			alive = false
		}()

		for alive {
			if commandSlice, err := collectCommandFiles(sessionPath, commandSet); err == nil {
				for _, commandFileName := range commandSlice {
					exitErr = processOperations(commandFileName, c, sessionPath, waitTime)

					if exitErr != nil {
						alive = false
					}

					commandSet.Insert(commandFileName)
				}
			} else {
				alive = false
				exitErr = err
			}
		}

		if exitErr == nil {
			// there is no scenario where the session just exits on its own. It's either an internal error,
			// an error caused by faulty commands,
			// or the default in this case a timeout
			exitErr = context.DeadlineExceeded
		}

		err := endSession(sessionPath, exitErr)
		return err
	}
}

/*
createLiveFolders

initializes all the folders where the data will be saved
*/
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

/*
RunLive

starts and runs a live session
*/
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

	err := chromedp.Run(ctx, tasks)

	if err != nil {
		log.Fatal(err)
	}
}
