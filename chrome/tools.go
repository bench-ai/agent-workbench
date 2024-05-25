/**
Key Terms:

snapshot: a snapshot is a name of a folder often passed in as a browser command argument.
It signifies what folder name you want to assign the data being collected too. It's useful because you can
group data collected at certain points of time together.


*/

package chrome

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileJob
/*
allows chromedp actions to write files in the background
*/
type FileJob struct {
	c  chan error
	wg sync.WaitGroup
}

// GetChannel
/*
get access to the error channel
*/
func (f *FileJob) GetChannel() chan error {
	return f.c
}

// GetWaitGroup
/*
get access to the wait group
*/
func (f *FileJob) GetWaitGroup() *sync.WaitGroup {
	return &f.wg
}

// InitFileJob
/*
creates a new file job struct
*/
func InitFileJob() *FileJob {
	return &FileJob{c: make(chan error)}
}

// writeBytes
/*
writes files to the given path, in the background
*/
func (f *FileJob) writeBytes(byteSlice []byte, writePath string) {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.c <- os.WriteFile(writePath, byteSlice, 0777)
	}()
}

// browserCommand
/*
interface for browser based actions

1) validate signifies if the struct has valid values

2) getAction returns the chromedp action func that can get executed
*/
type browserCommand interface {
	validate() error
	getAction(job *FileJob) chromedp.ActionFunc
}

// createSnapshotFolder
/*
creates the initial directory where the snapshot data will be saved
*/
func createSnapshotFolder(savePath, snapshot string) string {
	folderPath := filepath.Join(savePath, "snapshots", snapshot)
	imagePath := filepath.Join(folderPath, "images")
	if err := os.MkdirAll(imagePath, 0777); !os.IsExist(err) && err != nil {
		log.Fatal("Could not create directory: " + folderPath)
	}

	return folderPath
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

		cssMap := make([]map[string]string, 0)

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

// nodeMetaData
/*
carries all essential node data for web interaction
*/
type nodeMetaData struct {
	Id         int64               `json:"id"`
	Type       string              `json:"type"`
	Xpath      string              `json:"xpath"`
	Attributes map[string]string   `json:"attributes"`
	CssStyles  []map[string]string `json:"css_styles"`
}

// nodeWithStyles
/*
node with its correlating styles
*/
type nodeWithStyles struct {
	cssStyles []*css.ComputedStyleProperty
	node      *cdp.Node
}

// navigateToUrl
/*
opens a webpage with the url provided
*/
func navigateToUrl(
	url string) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.Navigate(url).Do(c)
		if err != nil {
			return err
		}
		return err
	}
}

// navigateWebPage
/*
handles url navigation
*/
type navigateWebPage struct {
	url string
}

// validate
/*
ensures the url provided is a valid url that can be navigated too
*/
func (n *navigateWebPage) validate() error {
	if !(strings.HasPrefix(n.url, "http://") || strings.HasPrefix(n.url, "https://")) {
		return errors.New("url must begin with http:// or https://")
	}
	return nil
}

// navInitFromJson
/*
turns json bytes into a navigateWebPage struct
*/
func navInitFromJson(jsonBytes []byte) *navigateWebPage {
	type body struct {
		Url string `json:"url"`
	}

	bdy := &body{}

	err := json.Unmarshal(jsonBytes, bdy)
	if err != nil {
		log.Fatal(err)
	}

	return &navigateWebPage{
		url: bdy.Url,
	}
}

// getAction
/*
converts the struct to ana action
*/
func (n *navigateWebPage) getAction(job *FileJob) chromedp.ActionFunc {
	_ = job
	return navigateToUrl(n.url)
}

// fullPageScreenShot
/*
struct that handles taking full page screenshots of the webpage on the browser

quality: how high resolution the image should be, higher is better

savePath: where the image should be saved
*/
type fullPageScreenShot struct {
	quality  uint8
	savePath string
}

// takeFullPageScreenshot
/*
takes screenshots of the entire webpage on the browser

quality: how high resolution the image should be, higher is better

buffer: a pointer to a slice that will store the data
*/
func takeFullPageScreenshot(
	quality uint8,
	buffer *[]byte) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.FullScreenshot(buffer, int(quality)).Do(c)
		if err != nil {
			return err
		}

		return err
	}
}

// fpsInitFromJson
/*
turns json bytes into a fpsInitFromJson struct
*/
func fpsInitFromJson(jsonBytes []byte, sessionPath string) *fullPageScreenShot {
	type body struct {
		Quality        uint8  `json:"quality"`
		Name           string `json:"name"`
		SnapShotFolder string `json:"snapshot_name"`
	}

	bdy := &body{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		log.Fatal(err)
	}

	folderPath := createSnapshotFolder(sessionPath, bdy.SnapShotFolder)

	fp := filepath.Join(folderPath, "images", bdy.Name)

	fullPage := fullPageScreenShot{
		savePath: fp,
		quality:  bdy.Quality,
	}

	return &fullPage
}

// validate
/*
ensures that properties of fpsInitFromJson such as quality and the savePath are valid
*/
func (f *fullPageScreenShot) validate() error {
	if f.quality <= 0 {
		return errors.New("quality must be greater than zero")
	}

	if !strings.HasSuffix(f.savePath, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	return nil
}

// getAction
/*
gives a chromedp action that collects a screenshot
*/
func (f *fullPageScreenShot) getAction(job *FileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var buffer []byte
		err := takeFullPageScreenshot(f.quality, &buffer).Do(c)
		if err == nil {
			job.writeBytes(buffer, f.savePath)
		}
		return err
	}
}

// takeElementScreenshot
/*
takes a screenshot of a dom element

scale: the size of the screenshot bigger scale larger picture, the picture tends to get more pixelated as the
size increases

selector: the xpath of the html element

buffer: a byte array pointer that will contain the image bytes
*/
func takeElementScreenshot(
	scale float64,
	selector string,
	buffer *[]byte) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.WaitVisible(selector).Do(c)
		if err != nil {
			return err
		}
		err = chromedp.ScreenshotScale(selector, scale, buffer, chromedp.NodeVisible).Do(c)
		if err != nil {
			return err
		}

		return err
	}
}

// elementScreenshot
/*
represents the values required to accurately screenshot a dom element
*/
type elementScreenshot struct {
	scale    float64
	selector string
	savePath string
}

// elemInitFromJson
/*
creates a elementScreenshot struct pointer from a json bytes

sessionPath is the location of the session folder
*/
func elemInitFromJson(jsonBytes []byte, sessionPath string) *elementScreenshot {
	type body struct {
		Scale          float64 `json:"scale"`
		Name           string  `json:"name"`
		Selector       string  `json:"selector"`
		SnapShotFolder string  `json:"snapshot_name"`
	}

	bdy := &body{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		log.Fatal(err)
	}

	folderPath := createSnapshotFolder(sessionPath, bdy.SnapShotFolder)

	folderPath = filepath.Join(folderPath, "images", bdy.Name)

	return &elementScreenshot{
		scale:    bdy.Scale,
		selector: bdy.Selector,
		savePath: folderPath,
	}
}

// validate
/*
ensures a elementScreenshot struct contains usable values
*/
func (e *elementScreenshot) validate() error {
	if !strings.HasSuffix(e.savePath, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	if e.scale <= 0 {
		return errors.New("scale must be greater than zero")
	}

	if e.selector == "" {
		return errors.New("selector cannot be empty")
	}

	return nil
}

// getAction
/*
returns a chromedp action that takes and write the screenshot
*/
func (e *elementScreenshot) getAction(job *FileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var buffer []byte
		err := takeElementScreenshot(e.scale, e.selector, &buffer).Do(c)
		if err == nil {
			job.writeBytes(buffer, e.savePath)
		}
		return err
	}
}

// contains all the data needed to click on an element
type click struct {
	selector  string
	queryFunc func(s *chromedp.Selector)
}

// clickOnElement
/*
Instructs the chrome agent to click on a section of the webpage

selector: data representing the elements location in the dom eg: xpath, js path, tag name

queryFunc: chromedp selector representing the query type eg: byId, xpath etc.
*/
func clickOnElement(
	selector string,
	queryFunc func(s *chromedp.Selector)) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.Click(selector, queryFunc).Do(c)
		return err
	}
}

// clickInitFromJson
/*
convert json bytes to click pointer
*/
func clickInitFromJson(jsonBytes []byte) (*click, error) {
	bdy := &struct {
		Selector  string `json:"selector"`
		QueryType string `json:"query_type"`
	}{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		return nil, err
	}

	c := click{
		selector: bdy.Selector,
	}

	switch bdy.QueryType {
	case "search":
		c.queryFunc = chromedp.BySearch
	default:
		return nil, fmt.Errorf("search type for click %v is not supported", bdy.QueryType)
	}

	return &c, nil
}

// validate
/*
ensures a click selector contains appropriate values
*/
func (c *click) validate() error {
	if c.selector == "" {
		return errors.New("click selector cannot be blank")
	}

	return nil
}

// getAction
/*
a chromedp action that clicks on an element
*/
func (c *click) getAction(job *FileJob) chromedp.ActionFunc {
	_ = job
	return clickOnElement(c.selector, c.queryFunc)
}

// sleep
/*
represents command that forces the browser to sleep for several ms amount of time
*/
type sleep struct {
	ms uint64
}

// sleep
/*
forces the browser to sleep for several ms amount of time
*/
func sleepForMs(ms uint64) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.Sleep(time.Duration(ms) * time.Millisecond).Do(c)
		if err != nil {
			log.Fatal(err)
		}

		return err
	}
}

// sleepInitFromJson
/*
initialize a sleep struct from json bytes
*/
func sleepInitFromJson(jsonBytes []byte) *sleep {
	bdy := &struct {
		Milliseconds uint64 `json:"ms"`
	}{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		log.Fatal(err)
	}

	return &sleep{
		ms: bdy.Milliseconds,
	}
}

// validate
/*
ensures the sleep command is valid (handled by the type system)
*/
func (s *sleep) validate() error {
	return nil
}

// getAction
/*
gets a chromedp command representing the action
*/
func (s *sleep) getAction(job *FileJob) chromedp.ActionFunc {
	_ = job
	return sleepForMs(s.ms)
}

// html
/*
struct representing a command that saves the html of the website being visited
*/
type html struct {
	savePath string
	selector string
}

// collectHtml
/*
collects the html present on the website
*/
func collectHtml(
	selector string,
	pString *string) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.OuterHTML(selector, pString).Do(c)
		if err != nil {
			log.Fatal(err)
		}

		return err
	}
}

// htmlInitFromJson
/*
turns json bytes into a html struct
*/
func htmlInitFromJson(jsonBytes []byte, sessionPath string) (*html, error) {
	type body struct {
		SnapShotFolder string  `json:"snapshot_name"`
		Selector       *string `json:"selector"`
	}

	bdy := &body{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		return nil, err
	}

	if bdy.SnapShotFolder == "" {
		return nil, errors.New(`html must be saved in a snapshot folder not name ""`)
	}

	savePath := createSnapshotFolder(sessionPath, bdy.SnapShotFolder)
	savePath = filepath.Join(savePath, "html.txt")

	var sel string

	if bdy.Selector == nil {
		sel = "html"
	} else {
		sel = *bdy.Selector
	}

	return &html{
		savePath: savePath,
		selector: sel,
	}, nil
}

// validate
/*
ensures html command contains valid elements
*/
func (h *html) validate() error {
	if h.selector == "" {
		return errors.New("selector cannot be blank")
	}

	return nil
}

// getAction
/*
a chromedp action that collects html
*/
func (h *html) getAction(job *FileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var text string
		err := collectHtml(h.selector, &text).Do(c)
		if err == nil {
			job.writeBytes([]byte(text), h.savePath)
		}
		return err
	}
}

// nodeCollect
/*
represents a command for collecting node data
*/
type nodeCollect struct {
	savePath      string
	selector      string
	prepopulate   bool
	recurse       bool
	collectStyles bool
}

// populateNode
/*
collects all node data

selector: collects only nodes belonging to this selector e.g: body

prepopulate: if true waits 1 second after requesting nodes (allows more to be collected)

recurse: if true uses bfs to collect all the child nodes of a parent node

collectStyles: if true collects all the css styles associated with a node

styledNodeList: the pointer that will hold all the data being collected
*/
func populatedNode(
	selector string,
	prepopulate bool,
	recurse bool,
	collectStyles bool,
	styledNodeList *[]*nodeWithStyles) chromedp.ActionFunc {
	return func(c context.Context) error {
		popSlice := make([]chromedp.PopulateOption, 0, 1)

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

		if recurse {
			nodeSlice = flattenNode(nodeSlice)
		}

		for _, node := range nodeSlice {

			var cs []*css.ComputedStyleProperty
			csErr := errors.New("failed")

			if collectStyles {
				cs, csErr = css.GetComputedStyleForNode(node.NodeID).Do(c)
			}

			styledNode := nodeWithStyles{
				node: node,
			}

			if csErr == nil {
				styledNode.cssStyles = cs
			}

			*styledNodeList = append(*styledNodeList, &styledNode)
		}

		return nil
	}
}

// nodeInitFromJson
/*
initializes node from json bytes
*/
func nodeInitFromJson(jsonBytes []byte, sessionPath string) *nodeCollect {
	type body struct {
		Selector       string `json:"selector"`
		GetStyles      bool   `json:"get_styles"`
		Prepopulate    bool   `json:"prepopulate"`
		Recurse        bool   `json:"recurse"`
		SnapShotFolder string `json:"snapshot_name"`
	}

	bdy := &body{}
	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		log.Fatal(err)
	}

	if bdy.SnapShotFolder == "" {
		log.Fatal("collect node method received a empty snapshot directory")
	}

	snapPath := createSnapshotFolder(sessionPath, bdy.SnapShotFolder)
	savePath := filepath.Join(snapPath, "nodes.json")

	nc := nodeCollect{
		selector:      bdy.Selector,
		savePath:      savePath,
		prepopulate:   bdy.Prepopulate,
		recurse:       bdy.Recurse,
		collectStyles: bdy.GetStyles,
	}

	return &nc
}

// validate
/*
ensures nodeCollect has valid values to initiate a command
*/
func (nc *nodeCollect) validate() error {
	if nc.selector == "" {
		return errors.New("selector is empty")
	}

	return nil
}

// getAction
/*
a chromedp action that collects all nodes
*/
func (nc *nodeCollect) getAction(job *FileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var nodeList []*nodeWithStyles
		err := populatedNode(
			nc.selector,
			nc.prepopulate,
			nc.recurse,
			nc.collectStyles,
			&nodeList).Do(c)

		if err != nil {
			return err
		}

		byteSlice, err := json.Marshal(parseThroughNodes(nodeList))

		if err != nil {
			log.Fatal("could not convert nodes to json bytes")
		}

		job.writeBytes(byteSlice, nc.savePath)

		return err
	}
}

//func (b *Executor) AcquireLocation(snapshot string) {
//	var loc string
//
//	b.appendTask(chromedp.ActionFunc(func(c context.Context) error {
//		err := chromedp.Location(&loc).Do(c)
//		if err != nil {
//			return err
//		}
//		return err
//	}))
//
//	b.locationMap[snapshot] = append(b.locationMap[snapshot], &loc)
//}
