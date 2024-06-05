package chrome

import (
	"context"
	"github.com/chromedp/chromedp"
	"strings"
	"testing"
	"time"
)

func getNavigateCommand() chromedp.ActionFunc {
	url := "https://bench-ai.com"
	return navigateToUrl(url)
}

func actionRunner(timeoutSeconds int64, actionFunc ...chromedp.ActionFunc) error {
	actx, _ := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false))...)

	ctx, _ := chromedp.NewContext(
		actx,
	)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)

	defer cancel()

	tasks := func() chromedp.Tasks {
		tskSlice := make(chromedp.Tasks, len(actionFunc))
		for index, task := range actionFunc {
			tskSlice[index] = task
		}
		return tskSlice
	}()

	return chromedp.Run(ctx, tasks)
}

func TestNavigate(t *testing.T) {
	var loc string

	getLocation := func(location *string) chromedp.ActionFunc {
		return func(ctx context.Context) error {
			return chromedp.Location(location).Do(ctx)
		}
	}

	err := actionRunner(3, getNavigateCommand(), getLocation(&loc))

	if !strings.Contains(loc, "bench-ai.com") {
		t.Fatal("navigate failed as invalid url was collected")
	}

	if err != nil {
		t.Fatal(err)
	}
}

func TestTakeFullPageScreenshot(t *testing.T) {
	var dataBuffer []byte
	action := takeFullPageScreenshot(2, &dataBuffer)
	err := actionRunner(3, getNavigateCommand(), action)

	if err != nil {
		t.Fatal(err)
	}

	if len(dataBuffer) == 0 {
		t.Fatal("screenshot was not taken, as no bytes have been collected")
	}
}

func TestTakeElementScreenshot(t *testing.T) {
	var dataBuffer []byte
	screenshot := takeElementScreenshot(1, "body", &dataBuffer)
	err := actionRunner(3, getNavigateCommand(), screenshot)

	if err != nil {
		t.Fatal(err)
	}

	if len(dataBuffer) == 0 {
		t.Fatal("screenshot was not taken, as no bytes have been collected")
	}
}

func TestSleep(t *testing.T) {
	start := time.Now()
	err := actionRunner(6, getNavigateCommand(), sleepForMs(1500))

	end := time.Now()

	if err != nil {
		t.Fatal(err)
	}

	if end.Sub(start).Milliseconds() <= 1500 {
		t.Fatal("did not sleep for requested duration")
	}
}

func TestNodeCollect(t *testing.T) {
	nav := getNavigateCommand()
	sle := sleepForMs(2000)

	var styledNodeList []*nodeWithStyles
	col := populatedNode("body", true, true, true, &styledNodeList)

	err := actionRunner(10, nav, sle, col)

	if err != nil {
		t.Fatal(err)
	}

	if len(styledNodeList) == 0 {
		t.Fatal("no nodes were collected")
	}

	var found bool

	for _, styledNodes := range styledNodeList {
		if len(styledNodes.cssStyles) > 0 {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("no node styles were collected")
	}
}

func TestClick(t *testing.T) {

	taskSlice := make([]chromedp.ActionFunc, 3)
	taskSlice[0] = getNavigateCommand()
	taskSlice[1] = sleepForMs(2000)

	var styledNodeList []*nodeWithStyles
	taskSlice[2] = populatedNode("body", true, true, true, &styledNodeList)

	err := actionRunner(10, taskSlice...)

	if err != nil {
		t.Fatal(err)
	}

	clk := clickOnElement(styledNodeList[0].node.FullXPath(), chromedp.BySearch)
	taskSlice[2] = clk

	err = actionRunner(10, taskSlice...)

	if err != nil {
		t.Fatal(err)
	}
}

func TestHtml(t *testing.T) {
	nav := getNavigateCommand()
	sle := sleepForMs(2000)
	var htmlString string

	colHtml := collectHtml("body", &htmlString)
	err := actionRunner(7, nav, sle, colHtml)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(htmlString, "<body") {
		t.Errorf("does not start the selected tag body")
	}

	if !strings.HasSuffix(htmlString, "/body>") {
		t.Errorf("does not end with the selected tag body")
	}
}

func TestScrollToPixel(t *testing.T) {
	nav := getNavigateCommand()
	sle := sleepForMs(2000)
	scroll := scrollToPixel(0, 500)
	err := actionRunner(7, nav, sle, scroll)

	if err != nil {
		t.Fatal(err)
	}
}

func TestScrollByPercentage(t *testing.T) {
	nav := getNavigateCommand()
	sle := sleepForMs(2000)
	err, scroll := scrollByPercentage(1)
	sle2 := sleepForMs(2000)

	if err != nil {
		t.Fatal(err)
	}

	err = actionRunner(7, nav, sle, scroll, sle2)
	if err != nil {
		t.Fatal(err)
	}
}
