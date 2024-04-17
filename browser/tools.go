package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"time"
)

type nodeMetaData struct {
	Id    int64  `json:"id"`
	Type  string `json:"type"`
	Xpath string `json:"xpath"`
}

type Executor struct {
	Url              string
	ctx              context.Context
	cancel           context.CancelFunc
	tasks            chromedp.Tasks
	imageList        []*[]byte
	fileNameList     []string
	snapshotNameList []string
	snapshotList     []*string
	nodeSavePath     string
	nodes            []*cdp.Node
}

func DeleteByIndex[T any](s []T, index int) (error, []T) {

	if index >= len(s) {
		return errors.New("index out of bounds"), nil
	}

	if index < 0 {
		return errors.New("index must be >= 0"), nil
	}

	if index == 0 {
		return nil, s[1:]
	}

	if index == len(s)-1 {
		return nil, s[:len(s)-1]
	}

	slice1 := s[:index]
	slice2 := s[index+1:]

	return nil, append(slice1, slice2...)
}

func (b *Executor) Init(headless bool) *Executor {

	/**
		ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	*/
	if headless {
		b.ctx, b.cancel = chromedp.NewContext(
			context.Background(),
		)
	} else {
		actx, _ := chromedp.NewExecAllocator(
			context.Background(),
			append(
				chromedp.DefaultExecAllocatorOptions[:],
				chromedp.Flag("headless", false))...)

		b.ctx, b.cancel = chromedp.NewContext(
			actx,
			chromedp.WithLogf(log.Printf),
			chromedp.WithErrorf(log.Printf))
	}

	return b
}

func (b *Executor) appendTask(action chromedp.Action) {
	b.tasks = append(b.tasks, action)
}

func (b *Executor) Navigate(url string) {
	b.tasks = append(b.tasks, chromedp.Navigate(url))
}

func (b *Executor) FullPageScreenShot(quality uint8, name string) {

	fmt.Println("here")
	var buf []byte
	b.appendTask(chromedp.FullScreenshot(&buf, int(quality)))
	b.imageList = append(b.imageList, &buf)
	b.fileNameList = append(b.fileNameList, name)
}

func (b *Executor) ElementScreenshot(scale float64, selector string, name string) {
	var buf []byte
	b.appendTask(chromedp.ScreenshotScale(selector, scale, &buf, chromedp.BySearch, chromedp.NodeVisible))
	b.imageList = append(b.imageList, &buf)
	b.fileNameList = append(b.fileNameList, name)
}

// SleepForSeconds
/*
Lets the browser pause operations for a certain amount of time
*/
func (b *Executor) SleepForSeconds(seconds int) {
	b.appendTask(
		chromedp.Sleep(time.Duration(seconds) * time.Second))
}

// SaveSnapshot
/*
Collects all the HTML associated with a webpage, saves all operations that led to the creation of the html,
we use it for snapshot purposes
*/
func (b *Executor) SaveSnapshot(snapshotName string) {
	/*
		TODO: Add instruction set to snapshot
	*/
	var snapShotHtml string
	b.snapshotNameList = append(b.snapshotNameList, snapshotName)
	b.snapshotList = append(b.snapshotList, &snapShotHtml)

	b.appendTask(chromedp.OuterHTML("body", &snapShotHtml))
}

// parseThroughNodes
/*
iterates through nodes and returns structures to recollect them
*/
func parseThroughNodes(nodeSlice []*cdp.Node) []nodeMetaData {

	deleteByIndex := DeleteByIndex[*cdp.Node]

	var nodeMetaDataSlice []nodeMetaData

	for len(nodeSlice) > 0 {
		for range nodeSlice {
			nodeSlice = append(nodeSlice, nodeSlice[0].Children...)

			metaData := nodeMetaData{
				Id:    nodeSlice[0].NodeID.Int64(),
				Type:  nodeSlice[0].NodeType.String(),
				Xpath: nodeSlice[0].FullXPath(),
			}

			nodeMetaDataSlice = append(nodeMetaDataSlice, metaData)
			_, nodeSlice = deleteByIndex(nodeSlice, 0)
		}
	}

	return nodeMetaDataSlice
}

// CollectNodes
/*
Collect all element nodes in the html webpage
*/
func (b *Executor) CollectNodes(selector, savePath string, waitReady bool) {

	if waitReady {
		b.appendTask(chromedp.WaitReady(selector))
	}

	fmt.Println(selector)
	b.appendTask(chromedp.Nodes(selector, &b.nodes))
	b.nodeSavePath = savePath
}

func (b *Executor) Execute() {
	defer b.cancel()
	if err := chromedp.Run(b.ctx, b.tasks); err != nil {
		log.Fatalf("Unable to run browser tasks due to: %v", err)
	}

	for index, buf := range b.imageList {
		path := filepath.Join("./resources", "images", b.fileNameList[index])
		if err := os.WriteFile(path, *buf, 0666); err != nil {
			log.Fatalf("Was unable to write file: %s, due to error: %v", path, err)
		}
	}

	for index, snapshotName := range b.snapshotNameList {

		folderPath := filepath.Join("./resources", "snapshots", snapshotName)
		if err := os.MkdirAll(folderPath, os.ModePerm); !os.IsExist(err) && err != nil {
			log.Fatal("Could not create directory: " + folderPath)
		}

		path := filepath.Join(folderPath, "WPBody.txt")

		byteSlice := []byte(*b.snapshotList[index])
		if err := os.WriteFile(path, byteSlice, 0666); err != nil {
			log.Fatalf("Was unable to write file: %s, due to error: %v", path, err)
		}
	}

	metaDataSlice := parseThroughNodes(b.nodes)
	byteSlice, err := json.Marshal(metaDataSlice)

	if err != nil {
		log.Fatalf("Unable to marshal node meta data: %v", err)
	}

	if err := os.WriteFile(b.nodeSavePath, byteSlice, 066); err != nil {
		log.Fatalf("Was unable to write file: %s, due to error: %v", b.nodeSavePath, err)
	}

	b.imageList = []*[]byte{}
	b.fileNameList = []string{}
	b.snapshotList = []*string{}
	b.snapshotNameList = []string{}
	b.nodes = []*cdp.Node{}
}
