package chrome

import (
	"context"
	"encoding/json"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

// AddOperation
/*
checks for if an operation exists and adds it to the execution queue
*/

func AddOperation(
	params map[string]interface{},
	commandName string,
	sessionPath string,
	job *FileJob) chromedp.Action {

	paramBytes, err := json.Marshal(params)

	if err != nil {
		log.Fatal("failed to marshall browser llm")
	}

	var browserParams browserCommand

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
		browserParams = clickInitFromJson(paramBytes)
	case "save_html":
		browserParams = htmlInitFromJson(paramBytes, sessionPath)
	case "sleep":
		browserParams = sleepInitFromJson(paramBytes)
	case "iterate_html":
		browserParams = htmlIterInitFromJson(paramBytes, sessionPath)
	default:
		log.Fatalf("%s is not a supported browser llm \n", commandName)
	}

	if err := browserParams.validate(); err != nil {
		log.Fatalf("%v", err)
	}

	return browserParams.getAction(job)
}

func runTasks(
	paramSlice []map[string]interface{},
	commandNameSlice []string,
	sessionPath string,
	ctx context.Context) {

	job := InitFileJob()
	tsk := make(chromedp.Tasks, len(commandNameSlice))

	for index, cName := range commandNameSlice {
		tsk[index] = AddOperation(paramSlice[index], cName, sessionPath, job)
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

func RunSequentialCommands(
	headless bool,
	timeout *int16,
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
