package browser

import (
	"context"
	"encoding/json"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"time"
)

type nodeMetaData struct {
	Id         int64               `json:"id"`
	Type       string              `json:"type"`
	Xpath      string              `json:"xpath"`
	Attributes map[string]string   `json:"attributes"`
	CssStyles  []map[string]string `json:"css_styles,omitempty"`
}

type nodeWithStyles struct {
	cssStyles []*css.ComputedStyleProperty
	node      *cdp.Node
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
	nodeMap   map[string]*[]*nodeWithStyles
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
				//chromedp.WindowSize(200, 200),
				chromedp.Flag("headless", false))...)

		b.ctx, b.cancel = chromedp.NewContext(
			actx,
			//chromedp.WithLogf(log.Printf),
			//chromedp.WithDebugf(log.Printf),
			//chromedp.WithErrorf(log.Printf))
		)
	}

	if timeout != nil {
		b.ctx, b.cancel = context.WithTimeout(b.ctx, time.Duration(*timeout)*time.Second)
	}

	b.htmlMap = make(map[string]*string)
	b.nodeMap = make(map[string]*[]*nodeWithStyles)
	b.imageList = make([]*imageMetaData, 0, 10)

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
	b.appendTask(chromedp.WaitVisible(selector))
	b.appendTask(chromedp.ScreenshotScale(selector, scale, &buf, chromedp.NodeVisible))

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
func parseThroughNodes(nodeSlice []*nodeWithStyles) []nodeMetaData {

	var nodeMetaDataSlice []nodeMetaData

	for _, nodeMd := range nodeSlice {
		attrMap := map[string]string{}

		for i := 0; i < len(nodeMd.node.Attributes); i += 2 {
			attrMap[nodeMd.node.Attributes[i]] += nodeMd.node.Attributes[i+1]
		}

		var cssMap []map[string]string

		for _, cascade := range nodeMd.cssStyles {
			tempMap := map[string]string{
				"name":  cascade.Name,
				"value": cascade.Value,
			}

			cssMap = append(cssMap, tempMap)
		}

		metaData := nodeMetaData{
			Id:         nodeMd.node.NodeID.Int64(),
			Type:       nodeMd.node.NodeType.String(),
			Xpath:      nodeMd.node.FullXPath(),
			Attributes: attrMap,
			CssStyles:  cssMap,
		}

		nodeMetaDataSlice = append(nodeMetaDataSlice, metaData)
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

func populatedNodeAction(
	selector string,
	prepopulate bool,
	recurse bool,
	nodesWithStyles *[]*nodeWithStyles) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(c context.Context) error {
			var popSlice []chromedp.PopulateOption
			if prepopulate {
				popSlice = append(popSlice, chromedp.PopulateWait(1*time.Second))
			}

			var nodeSlice []*cdp.Node

			err := chromedp.Nodes(
				selector,
				&nodeSlice,
				chromedp.Populate(-1, true, popSlice...),
			).Do(c)

			if err != nil {
				return err
			}

			if nodesWithStyles != nil {

				if recurse {
					nodeSlice = flattenNode(nodeSlice)
				}

				for _, node := range nodeSlice {
					if cs, err := css.GetComputedStyleForNode(node.NodeID).Do(c); err == nil {
						*nodesWithStyles = append(*nodesWithStyles, &nodeWithStyles{cs, node})
					} else {
						*nodesWithStyles = append(*nodesWithStyles, &nodeWithStyles{node: node})
					}
				}
			}

			return nil
		}),
	}
}

// CollectNodes
/*
Collect all element nodes in the html webpage with the options to add css styles
*/
func (b *Executor) CollectNodes(
	selector,
	snapshotName string,
	prepopulate,
	waitReady,
	recurse,
	nodesWithStyles bool,

) {

	nodeSlice := make([]*nodeWithStyles, 0, 100)

	if waitReady {
		b.appendTask(chromedp.WaitReady(selector))
	}

	b.appendTask(
		populatedNodeAction(
			selector,
			prepopulate,
			recurse,
			func() *[]*nodeWithStyles {
				if nodesWithStyles {
					return &nodeSlice
				} else {
					return nil
				}
			}()),
	)

	b.nodeMap[snapshotName] = &nodeSlice
}

func (b *Executor) HtmlIterator(
	iterLimit uint16,
	pauseTime uint32,
	startingSnapshot uint8,
	snapshotName string,
	imageQuality uint8,
	saveImg bool,
	saveHtml bool,
	saveNodes bool,
) {

	pImgList := &b.imageList
	if !saveImg {
		pImgList = nil
	}

	pHtmlMap := b.htmlMap
	if !saveHtml {
		pHtmlMap = nil
	}

	pNodeMap := b.nodeMap
	if !saveNodes {
		pHtmlMap = nil
	}

	b.appendTask(
		htmlIteratorAction(
			iterLimit, pauseTime, startingSnapshot, snapshotName, imageQuality, pHtmlMap, pNodeMap, pImgList,
		),
	)
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
	b.nodeMap = map[string]*[]*nodeWithStyles{}
}
