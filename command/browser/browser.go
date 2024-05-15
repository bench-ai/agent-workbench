package browser

import (
	"agent/chrome"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
	"time"
)

type Params interface {
	Validate() error
}

type BrowserParams interface {
	Params
	AppendTask(b *browser.Executor)
}

type FullPageScreenShot struct {
	Quality        uint8  `json:"quality"`
	Name           string `json:"name"`
	SnapShotFolder string `json:"snapshot_name"`
}

func (f *FullPageScreenShot) Validate() error {
	if f.Quality == 0 {
		return errors.New("quality must be greater than zero")
	}

	if !strings.HasSuffix(f.Name, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	return nil
}

func (f *FullPageScreenShot) AppendTask(b *browser.Executor) {
	b.FullPageScreenShot(f.Quality, f.Name, f.SnapShotFolder)
}

type OpenWebPage struct {
	Url string `json:"url"`
}

func (o *OpenWebPage) Validate() error {
	if !(strings.HasPrefix(o.Url, "http://") || strings.HasPrefix(o.Url, "https://")) {
		return errors.New("url must begin with http:// or https://")
	}

	return nil
}

func (o *OpenWebPage) AppendTask(b *browser.Executor) {
	b.Navigate(o.Url)
}

type ElementScreenshot struct {
	Scale          float64 `json:"scale"`
	Name           string  `json:"name"`
	Selector       string  `json:"selector"`
	SnapShotFolder string  `json:"snapshot_name"`
}

func (e *ElementScreenshot) Validate() error {
	if !strings.HasSuffix(e.Name, ".jpg") {
		return errors.New("name must end with .jpg")
	}

	if e.Scale < 0 {
		return errors.New("scale must be greater than zero")
	}

	return nil
}

func (e *ElementScreenshot) AppendTask(b *browser.Executor) {
	b.ElementScreenshot(e.Scale, e.Selector, e.Name, e.SnapShotFolder)
}

type CollectNodes struct {
	Selector       string `json:"selector"`
	WaitReady      bool   `json:"wait_ready"`
	GetStyles      bool   `json:"get_styles"`
	Prepopulate    bool   `json:"prepopulate"`
	Recurse        bool   `json:"recurse"`
	SnapShotFolder string `json:"snapshot_name"`
}

func (c *CollectNodes) Validate() error {
	if strings.Contains(c.SnapShotFolder, ".") {
		return errors.New("snapshot_folder must be folder not a file")
	}
	return nil
}

func (c *CollectNodes) AppendTask(b *browser.Executor) {
	b.CollectNodes(c.Selector, c.SnapShotFolder, c.Prepopulate, c.WaitReady, c.Recurse, c.GetStyles)
}

type Click struct {
	Selector  string `json:"selector"`
	QueryType string `json:"query_type"`
}

func (c *Click) Validate() error {

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

func (c *Click) AppendTask(b *browser.Executor) {
	var query func(s *chromedp.Selector)

	switch c.QueryType {
	case "search":
		query = chromedp.BySearch
	default:
		log.Fatalf("unspported querytype %s", c.QueryType)
	}

	b.Click(c.Selector, query)
}

type SaveHtml struct {
	SnapShotFolder string `json:"snapshot_name"`
}

func (s *SaveHtml) Validate() error {
	if strings.Contains(s.SnapShotFolder, ".") {
		return errors.New("snapshot_folder must be folder not a file")
	}
	return nil
}

func (s *SaveHtml) AppendTask(b *browser.Executor) {
	b.SaveSnapshot(s.SnapShotFolder)
}

type Sleep struct {
	Seconds uint16 `json:"seconds"`
}

func (s *Sleep) Validate() error {
	return nil
}

func (s *Sleep) AppendTask(b *browser.Executor) {
	b.SleepForSeconds(s.Seconds)
}

type IterateHtml struct {
	IterLimit         *uint16 `json:"iter_limit"`
	PauseTime         *uint32 `json:"pause_time"`
	StartingSnapshot  *uint8  `json:"starting_snapshot"`
	SnapshotName      string  `json:"snapshot_name"`
	SaveHtml          bool    `json:"save_html"`
	SaveNode          bool    `json:"save_node"`
	SaveFullPageImage bool    `json:"save_full_page_image"`
	ImageQuality      uint8   `json:"image_quality"`
}

func (i *IterateHtml) Validate() error {

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

func (i *IterateHtml) AppendTask(b *browser.Executor) {

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

type AcquireLocation struct {
	SnapShotFolder string `json:"snapshot_name"`
}

func (a *AcquireLocation) Validate() error {
	if strings.Contains(a.SnapShotFolder, ".") {
		return errors.New("snapshot_folder must be folder not a file")
	}
	return nil
}

func (a *AcquireLocation) AppendTask(b *browser.Executor) {
	b.AcquireLocation(a.SnapShotFolder)
}
