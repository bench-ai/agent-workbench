/**
contains methods to help create and execute browser commands
*/

package chrome

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

// AddOperation
/*
checks for if an operation exists and converts it to a chromedp action
*/
func AddOperation(
	params map[string]interface{},
	commandName string,
	sessionPath string,
	job *FileJob) (chromedp.Action, error) {

	paramBytes, err := json.Marshal(params)

	if err != nil {
		return nil, err
	}

	var browserParams browserCommand
	var browserError error

	switch commandName {
	case "open_web_page":
		browserParams = navInitFromJson(paramBytes)
	case "full_page_screenshot":
		browserParams = fpsInitFromJson(paramBytes, sessionPath)
	case "element_screenshot":
		browserParams = elemInitFromJson(paramBytes, sessionPath)
	case "collect_nodes":
		browserParams = nodeInitFromJson(paramBytes, sessionPath)
	case "click":
		browserParams, browserError = clickInitFromJson(paramBytes)
		if browserError != nil {
			return nil, err
		}
	case "save_html":
		browserParams, browserError = htmlInitFromJson(paramBytes, sessionPath)
		if browserError != nil {
			return nil, err
		}
	case "sleep":
		browserParams = sleepInitFromJson(paramBytes)
	case "iterate_html":
		browserParams = htmlIterInitFromJson(paramBytes, sessionPath)
	default:
		return nil, fmt.Errorf("%s is not a supported browser llm \n", commandName)
	}

	if err := browserParams.validate(); err != nil {
		return nil, err
	}

	return browserParams.getAction(job), nil
}

// runTasks
/*
converts a map to chrome dp actions and executes them
*/
func runTasks(
	paramSlice []map[string]interface{},
	commandNameSlice []string,
	sessionPath string,
	ctx context.Context) {

	job := InitFileJob()
	tsk := make(chromedp.Tasks, len(commandNameSlice))

	for index, cName := range commandNameSlice {
		act, err := AddOperation(paramSlice[index], cName, sessionPath, job)
		if err != nil {
			log.Fatal(err)
		}
		tsk[index] = act
	}

	err := chromedp.Run(ctx, tsk)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		job.wg.Wait()
		close(job.c)
	}()

	for err = range job.c {
		if err != nil {
			log.Fatal(err)
		}
	}
}

// RunSequentialCommands
/*
runs commands sequentially based on the order provided, and initializes the requested chromedp context
*/
func RunSequentialCommands(
	headless bool,
	timeout *int32,
	sessionPath string,
	paramSlice []map[string]interface{},
	commandNameSlice []string,
) {
	var ctx context.Context
	var cancel context.CancelFunc

	if headless {
		ctx, cancel = chromedp.NewContext(
			context.Background(),
		)
	} else {
		actx, _ := chromedp.NewExecAllocator(
			context.Background(),
			append(
				chromedp.DefaultExecAllocatorOptions[:],
				chromedp.Flag("headless", false))...)

		ctx, cancel = chromedp.NewContext(
			actx,
		)
	}

	if timeout != nil {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*timeout)*time.Second)
	}

	defer cancel()

	runTasks(paramSlice, commandNameSlice, sessionPath, ctx)
}
