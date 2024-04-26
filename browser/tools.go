package browser

import (
	"agent/helper"
	"context"
	"encoding/json"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"time"
)

type nodeMetaData struct {
	Id         int64             `json:"id"`
	Type       string            `json:"type"`
	Xpath      string            `json:"xpath"`
	Attributes map[string]string `json:"attributes"`
}

type imageMetaData struct {
	snapShotName string
	imageName    string
	byteData     *[]byte
}

type Executor struct {
	Url       string
	ctx       context.Context
	cancel    context.CancelFunc
	tasks     chromedp.Tasks
	imageList []*imageMetaData
	htmlMap   map[string]*string
	nodeMap   map[string]*[]*cdp.Node
}

func (b *Executor) Init(headless bool, timeout *int16) *Executor {

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

	if timeout != nil {
		b.ctx, b.cancel = context.WithTimeout(b.ctx, time.Duration(*timeout)*time.Second)
	}

	b.htmlMap = map[string]*string{}
	b.nodeMap = map[string]*[]*cdp.Node{}

	return b
}

func (b *Executor) appendTask(action chromedp.Action) {
	b.tasks = append(b.tasks, action)
}

func (b *Executor) Navigate(url string) {
	b.tasks = append(b.tasks, chromedp.Navigate(url))
}

func (b *Executor) FullPageScreenShot(quality uint8, name, snapshot string) {
	var buf []byte
	var imageData imageMetaData
	b.appendTask(chromedp.FullScreenshot(&buf, int(quality)))

	imageData.byteData = &buf
	imageData.snapShotName = snapshot
	imageData.imageName = name

	b.imageList = append(b.imageList, &imageData)
}

func (b *Executor) ElementScreenshot(scale float64, selector string, name, snapshot string) {
	var buf []byte
	var imageData imageMetaData

	b.appendTask(chromedp.ScreenshotScale(selector, scale, &buf, chromedp.BySearch, chromedp.NodeVisible))

	imageData.byteData = &buf
	imageData.snapShotName = snapshot
	imageData.imageName = name

	b.imageList = append(b.imageList, &imageData)
}

// Click
/*
Instructs the browser agent to click on a section of the webpage
*/
func (b *Executor) Click(selector string, queryFunc func(s *chromedp.Selector)) {
	b.appendTask(chromedp.Click(selector, queryFunc))
}

// SleepForSeconds
/*
Lets the browser pause operations for a certain amount of time
*/
func (b *Executor) SleepForSeconds(seconds uint16) {
	b.appendTask(
		chromedp.Sleep(time.Duration(seconds) * time.Second))
}

// SaveSnapshot
/*
Collects all the HTML associated with a webpage, saves all operations that led to the creation of the html,
we use it for snapshot purposes
*/
func (b *Executor) SaveSnapshot(snapshotName string) {
	var snapShotHtml string
	b.appendTask(chromedp.OuterHTML("body", &snapShotHtml))
	b.htmlMap[snapshotName] = &snapShotHtml
}

// parseThroughNodes
/*
iterates through nodes and returns structures to recollect them
*/
func parseThroughNodes(nodeSlice []*cdp.Node) []nodeMetaData {

	deleteByIndex := helper.DeleteByIndex[*cdp.Node]

	var nodeMetaDataSlice []nodeMetaData

	for len(nodeSlice) > 0 {
		for range nodeSlice {
			nodeSlice = append(nodeSlice, nodeSlice[0].Children...)

			attrMap := map[string]string{}

			for i := 0; i < len(nodeSlice[0].Attributes); i += 2 {
				attrMap[nodeSlice[0].Attributes[i]] += nodeSlice[0].Attributes[i+1]
			}

			metaData := nodeMetaData{
				Id:         nodeSlice[0].NodeID.Int64(),
				Type:       nodeSlice[0].NodeType.String(),
				Xpath:      nodeSlice[0].FullXPath(),
				Attributes: attrMap,
			}

			nodeMetaDataSlice = append(nodeMetaDataSlice, metaData)
			_, nodeSlice = deleteByIndex(nodeSlice, 0)
		}
	}

	return nodeMetaDataSlice
}

// createSnapshotFolder
/*
Creates Snapshot folder if it does not exist already
*/
func createSnapshotFolder(snapshot string) string {
	folderPath := filepath.Join("./resources", "snapshots", snapshot)
	imagePath := filepath.Join(folderPath, "images")
	if err := os.MkdirAll(imagePath, os.ModePerm); !os.IsExist(err) && err != nil {
		log.Fatal("Could not create directory: " + folderPath)
	}

	return folderPath
}

// CollectNodes
/*
Collect all element nodes in the html webpage
*/
func (b *Executor) CollectNodes(selector, snapshotName string, waitReady bool) {
	var nodeSlice []*cdp.Node

	if waitReady {
		b.appendTask(chromedp.WaitReady(selector))
	}

	b.appendTask(chromedp.Nodes(selector, &nodeSlice))
	b.nodeMap[snapshotName] = &nodeSlice
}

func (b *Executor) Execute() {
	defer b.cancel()
	if err := chromedp.Run(b.ctx, b.tasks); err != nil {
		log.Fatalf("Unable to run browser tasks due to: %v", err)
	}

	for _, imd := range b.imageList {

		folderPath := createSnapshotFolder(imd.snapShotName)
		path := filepath.Join(folderPath, "images", imd.imageName)
		if err := os.WriteFile(path, *imd.byteData, 0666); err != nil {
			log.Fatalf("Was unable to write file: %s, due to error: %v", path, err)
		}
	}

	for snapShotName, html := range b.htmlMap {

		folderPath := createSnapshotFolder(snapShotName)

		path := filepath.Join(folderPath, "body.txt")

		byteSlice := []byte(*html)
		if err := os.WriteFile(path, byteSlice, 0666); err != nil {
			log.Fatalf("Was unable to write file: %s, due to error: %v", path, err)
		}
	}

	for snapShotName, node := range b.nodeMap {
		folderPath := createSnapshotFolder(snapShotName)

		path := filepath.Join(folderPath, "nodeData.json")

		metaDataSlice := parseThroughNodes(*node)

		byteSlice, err := json.MarshalIndent(metaDataSlice, "", "    ")

		if err != nil {
			log.Fatalf("Unable to marshal node meta data: %v", err)
		}

		if err := os.WriteFile(path, byteSlice, 0666); err != nil {
			log.Fatalf("Was unable to write file: %s, due to error: %v", path, err)
		}
	}

	b.imageList = []*imageMetaData{}
	b.htmlMap = map[string]*string{}
	b.nodeMap = map[string]*[]*cdp.Node{}
}
