package chrome

import (
	"context"
	"github.com/chromedp/chromedp"
	"os"
	"testing"
	"time"
)

func taskData() ([3]string, [3]map[string]interface{}) {
	commandNameSlice := [3]string{
		"open_web_page",
		"full_page_screenshot",
		"save_html",
	}

	paramSlice := [3]map[string]interface{}{
		{
			"url": "https://bench-ai.com",
		},
		{
			"quality":       5,
			"name":          "test.jpg",
			"snapshot_name": "s1",
		},
		{
			"snapshot_name": "s1",
			"selector":      "body",
		},
	}

	return commandNameSlice, paramSlice
}

func TestRunTasks(t *testing.T) {
	actx, _ := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true))...)

	ctx, _ := chromedp.NewContext(
		actx,
	)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(10)*time.Second)

	commandName, paramMap := taskData()

	er, td := tempDir()

	if er != nil {
		t.Fatal(er)
	}

	defer cancel()
	defer func() {
		if err := os.RemoveAll(td); err != nil {
			t.Fatal(err)
		}
	}()

	runTasks(paramMap[:], commandName[:], td, ctx)
}
