package command

import (
	"agent/browser"
	"errors"
	"strings"
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
	SnapShotFolder string `json:"snap_shot_name"`
}

func (f *FullPageScreenShot) Validate() error {
	if f.Quality == 0 {
		return errors.New("quality must be greater than zero")
	}

	if !strings.HasSuffix(f.Name, ".png") {
		return errors.New("name must end with .png")
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
	SnapShotFolder string  `json:"snap_shot_name"`
}

func (e *ElementScreenshot) Validate() error {
	if !strings.HasSuffix(e.Name, ".png") {
		return errors.New("name must end with .png")
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
	SnapShotFolder string `json:"snap_shot_name"`
}

func (c *CollectNodes) Validate() error {
	if strings.Contains(c.SnapShotFolder, ".") {
		return errors.New("snap_shot_folder must be folder not a file")
	}
	return nil
}

func (c *CollectNodes) AppendTask(b *browser.Executor) {
	b.CollectNodes(c.Selector, c.SnapShotFolder, c.WaitReady)
}
