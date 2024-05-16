package chrome

import (
	"context"
	"encoding/json"
	"errors"
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

type fileJob struct {
	c  chan error
	wg sync.WaitGroup
}

func initFileJob() *fileJob {
	return &fileJob{c: make(chan error)}
}

func (f *fileJob) writeBytes(byteSlice []byte, writePath string) {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.c <- os.WriteFile(writePath, byteSlice, 0777)
	}()
}

type browserCommand interface {
	validate() error
	getAction(job *fileJob) chromedp.ActionFunc
}

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

type nodeMetaData struct {
	Id         int64               `json:"id"`
	Type       string              `json:"type"`
	Xpath      string              `json:"xpath"`
	Attributes map[string]string   `json:"attributes"`
	CssStyles  []map[string]string `json:"css_styles"`
}

type nodeWithStyles struct {
	cssStyles []*css.ComputedStyleProperty
	node      *cdp.Node
}

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

type navigateWebPage struct {
	url string
}

func (n *navigateWebPage) validate() error {
	if !(strings.HasPrefix(n.url, "http://") || strings.HasPrefix(n.url, "https://")) {
		return errors.New("url must begin with http:// or https://")
	}
	return nil
}

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

func (n *navigateWebPage) getAction(job *fileJob) chromedp.ActionFunc {
	_ = job
	return navigateToUrl(n.url)
}

type fullPageScreenShot struct {
	quality  uint8
	savePath string
}

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

func (f *fullPageScreenShot) validate() error {
	if f.quality <= 0 {
		return errors.New("quality must be greater than zero")
	}

	if !strings.HasSuffix(f.savePath, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	return nil
}

func (f *fullPageScreenShot) getAction(job *fileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var buffer []byte
		err := takeFullPageScreenshot(f.quality, &buffer).Do(c)
		if err == nil {
			job.writeBytes(buffer, f.savePath)
		}
		return err
	}
}

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

type elementScreenshot struct {
	scale    float64
	selector string
	savePath string
}

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

func (e *elementScreenshot) validate() error {
	if !strings.HasSuffix(e.savePath, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	if e.scale <= 0 {
		return errors.New("scale must be greater than zero")
	}

	return nil
}

func (e *elementScreenshot) getAction(job *fileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var buffer []byte
		err := takeElementScreenshot(e.scale, e.selector, &buffer).Do(c)
		if err == nil {
			job.writeBytes(buffer, e.savePath)
		}
		return err
	}
}

type click struct {
	selector  string
	queryFunc func(s *chromedp.Selector)
}

// clickOnElement
/*
Instructs the chrome agent to click on a section of the webpage
*/
func clickOnElement(
	selector string,
	queryFunc func(s *chromedp.Selector)) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.Click(selector, queryFunc).Do(c)
		return err
	}
}

func clickInitFromJson(jsonBytes []byte) *click {
	bdy := &struct {
		Selector  string `json:"selector"`
		QueryType string `json:"query_type"`
	}{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		log.Fatal(err)
	}

	c := click{
		selector: bdy.Selector,
	}

	switch bdy.QueryType {
	case "search":
		c.queryFunc = chromedp.BySearch
	default:
		log.Fatalf("search type for click %v is not supported", bdy.QueryType)
	}

	return &c
}

func (c *click) validate() error {
	if c.selector == "" {
		return errors.New("click selector cannot be blank")
	}

	return nil
}

func (c *click) getAction(job *fileJob) chromedp.ActionFunc {
	_ = job
	return clickOnElement(c.selector, c.queryFunc)
}

type sleep struct {
	ms uint64
}

func sleepForMs(ms uint64) chromedp.ActionFunc {
	return func(c context.Context) error {
		err := chromedp.Sleep(time.Duration(ms) * time.Millisecond).Do(c)
		if err != nil {
			log.Fatal(err)
		}

		return err
	}
}

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

func (s *sleep) validate() error {
	return nil
}

func (s *sleep) getAction(job *fileJob) chromedp.ActionFunc {
	_ = job
	return sleepForMs(s.ms)
}

type html struct {
	savePath string
	selector string
}

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

func htmlInitFromJson(jsonBytes []byte, sessionPath string) *html {
	type body struct {
		SnapShotFolder string  `json:"snapshot_name"`
		Selector       *string `json:"selector"`
	}

	bdy := &body{}

	err := json.Unmarshal(jsonBytes, bdy)

	if err != nil {
		log.Fatal(err)
	}

	if bdy.SnapShotFolder == "" {
		log.Fatal(`html must be saved in a snapshot folder not name ""`)
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
	}
}

func (h *html) validate() error {
	return nil
}

func (h *html) getAction(job *fileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var text string
		err := collectHtml(h.selector, &text).Do(c)
		if err == nil {
			job.writeBytes([]byte(text), h.savePath)
		}
		return err
	}
}

type nodeCollect struct {
	savePath      string
	selector      string
	prepopulate   bool
	recurse       bool
	collectStyles bool
}

func populatedNode(
	selector string,
	prepopulate bool,
	recurse bool,
	collectStyles bool,
	styledNodeList *[]*nodeWithStyles) chromedp.ActionFunc {
	return func(c context.Context) error {
		popSlice := make([]chromedp.PopulateOption, 0, 1)

		// Add external node access to all file methods
		// Add save request

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

func (nc *nodeCollect) validate() error {
	if nc.selector == "" {
		return errors.New("selector is empty")
	}

	return nil
}

func (nc *nodeCollect) getAction(job *fileJob) chromedp.ActionFunc {
	return func(c context.Context) error {
		var nodeList []*nodeWithStyles
		err := populatedNode(
			nc.selector,
			nc.prepopulate,
			nc.recurse,
			nc.collectStyles,
			&nodeList).Do(c)

		if err != nil {
			log.Fatal(err)
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
