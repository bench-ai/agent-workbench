package browser

import (
	"agent/chrome"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
	"time"
)

type params interface {
	Validate() error
}

type appendParams interface {
	params
	AppendTask(b *chrome.Executor)
}

type fullPageScreenShot struct {
	Quality        uint8  `json:"quality"`
	Name           string `json:"name"`
	SnapShotFolder string `json:"snapshot_name"`
}

func (f *fullPageScreenShot) Validate() error {
	if f.Quality == 0 {
		return errors.New("quality must be greater than zero")
	}

	if !strings.HasSuffix(f.Name, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	return nil
}

func (f *fullPageScreenShot) AppendTask(b *chrome.Executor) {
	b.FullPageScreenShot(f.Quality, f.Name, f.SnapShotFolder)
}

type openWebPage struct {
	Url string `json:"url"`
}

func (o *openWebPage) Validate() error {
	if !(strings.HasPrefix(o.Url, "http://") || strings.HasPrefix(o.Url, "https://")) {
		return errors.New("url must begin with http:// or https://")
	}

	return nil
}

func (o *openWebPage) AppendTask(b *chrome.Executor) {
	b.Navigate(o.Url)
}

type elementScreenshot struct {
	Scale          float64 `json:"scale"`
	Name           string  `json:"name"`
	Selector       string  `json:"selector"`
	SnapShotFolder string  `json:"snapshot_name"`
}

func (e *elementScreenshot) Validate() error {
	if !strings.HasSuffix(e.Name, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	if e.Scale < 0 {
		return errors.New("scale must be greater than zero")
	}

	return nil
}

func (e *elementScreenshot) AppendTask(b *chrome.Executor) {
	b.ElementScreenshot(e.Scale, e.Selector, e.Name, e.SnapShotFolder)
}

type collectNodes struct {
	Selector       string `json:"selector"`
	WaitReady      bool   `json:"wait_ready"`
	GetStyles      bool   `json:"get_styles"`
	Prepopulate    bool   `json:"prepopulate"`
	Recurse        bool   `json:"recurse"`
	SnapShotFolder string `json:"snapshot_name"`
}

func (c *collectNodes) Validate() error {
	if strings.Contains(c.SnapShotFolder, ".") {
		return errors.New("snapshot_folder must be folder not a file")
	}
	return nil
}

func (c *collectNodes) AppendTask(b *chrome.Executor) {
	b.CollectNodes(c.Selector, c.SnapShotFolder, c.Prepopulate, c.WaitReady, c.Recurse, c.GetStyles)
}

type click struct {
	Selector  string `json:"selector"`
	QueryType string `json:"query_type"`
}

func (c *click) Validate() error {

	if c.Selector == "" {
		return errors.New("selector is required")
	}

	validTypes := [1]string{
		"search",
	}

	for _, i := range validTypes {
		if c.QueryType == i {
			return nil
		}
	}

	return fmt.Errorf("query type %s not supported", c.QueryType)
}

func (c *click) AppendTask(b *chrome.Executor) {
	var query func(s *chromedp.Selector)

	switch c.QueryType {
	case "search":
		query = chromedp.BySearch
	default:
		log.Fatalf("unspported querytype %s", c.QueryType)
	}

	b.Click(c.Selector, query)
}

type saveHtml struct {
	SnapShotFolder string `json:"snapshot_name"`
}

func (s *saveHtml) Validate() error {
	if strings.Contains(s.SnapShotFolder, ".") {
		return errors.New("snapshot_folder must be folder not a file")
	}
	return nil
}

func (s *saveHtml) AppendTask(b *chrome.Executor) {
	b.SaveSnapshot(s.SnapShotFolder)
}

type sleep struct {
	Seconds uint16 `json:"seconds"`
}

func (s *sleep) Validate() error {
	return nil
}

func (s *sleep) AppendTask(b *chrome.Executor) {
	b.SleepForSeconds(s.Seconds)
}

type iterateHtml struct {
	IterLimit         *uint16 `json:"iter_limit"`
	PauseTime         *uint32 `json:"pause_time"`
	StartingSnapshot  *uint8  `json:"starting_snapshot"`
	SnapshotName      string  `json:"snapshot_name"`
	SaveHtml          bool    `json:"save_html"`
	SaveNode          bool    `json:"save_node"`
	SaveFullPageImage bool    `json:"save_full_page_image"`
	ImageQuality      uint8   `json:"image_quality"`
}

func (i *iterateHtml) Validate() error {

	if i.PauseTime == nil {
		pause := uint32(500)
		i.PauseTime = &pause
	}

	if i.IterLimit == nil {
		iterCount := 10 * time.Minute / (time.Duration(*i.PauseTime) * time.Millisecond)
		iterCountCeil := uint16(iterCount.Minutes())
		i.IterLimit = &iterCountCeil
	}

	if i.StartingSnapshot == nil {
		ss := uint8(0)
		i.StartingSnapshot = &ss
	}

	if i.SnapshotName == "" {
		return errors.New("snapshot name for iterate html is blank")
	}

	if i.SaveFullPageImage {
		if i.ImageQuality == 0 {
			return errors.New("image quality must be provided and greater than 0")
		}
	}

	return nil
}

func (i *iterateHtml) AppendTask(b *chrome.Executor) {

	b.HtmlIterator(
		*i.IterLimit,
		*i.PauseTime,
		*i.StartingSnapshot,
		i.SnapshotName,
		i.ImageQuality,
		i.SaveFullPageImage,
		i.SaveHtml,
		i.SaveNode)
}

type acquireLocation struct {
	SnapShotFolder string `json:"snapshot_name"`
}

func (a *acquireLocation) Validate() error {
	if strings.Contains(a.SnapShotFolder, ".") {
		return errors.New("snapshot_folder must be folder not a file")
	}
	return nil
}

func (a *acquireLocation) AppendTask(b *chrome.Executor) {
	b.AcquireLocation(a.SnapShotFolder)
}

// addOperation
/*
checks for if an operation exists and adds it to the execution queue
*/
func AddOperation(params map[string]interface{}, commandName string, builder *chrome.Executor) {

	paramBytes, _ := json.Marshal(params)
	var browserParams appendParams

	switch commandName {
	case "open_web_page":
		browserParams = &openWebPage{}
	case "full_page_screenshot":
		browserParams = &fullPageScreenShot{}
	case "element_screenshot":
		browserParams = &elementScreenshot{}
	case "collect_nodes":
		browserParams = &collectNodes{}
	case "click":
		browserParams = &click{}
	case "save_html":
		browserParams = &saveHtml{}
	case "sleep":
		browserParams = &sleep{}
	case "iterate_html":
		browserParams = &iterateHtml{}
	case "acquire_location":
		browserParams = &acquireLocation{}
	default:
		log.Fatalf("%s is not a supported chrome command \n", commandName)
	}

	if err := json.Unmarshal(paramBytes, browserParams); err != nil {
		log.Fatalf("failed to parse %s command \n", commandName)
	}

	if err := browserParams.Validate(); err != nil {
		log.Fatalf("%v", err)
	}

	browserParams.AppendTask(builder)
}
