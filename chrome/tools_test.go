package chrome

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func createArtificialNodes(id int64, children []*cdp.Node) *nodeWithStyles {

	attr := [2]string{
		"test", "test1",
	}

	localNames := [5]string{
		"img", "body", "html", "div", "p",
	}

	local := localNames[rand.Intn(5)]

	pNode := &cdp.Node{
		NodeID:     cdp.NodeID(id),
		NodeType:   cdp.NodeType(rand.Intn(12) + 1),
		Attributes: attr[:],
		LocalName:  local,
		Children:   children,
	}

	fStyle := css.ComputedStyleProperty{
		Name:  "test",
		Value: "value",
	}

	masterStyle := &nodeWithStyles{
		[]*css.ComputedStyleProperty{
			&fStyle,
		},
		pNode,
	}

	return masterStyle
}

func buildNodeTree() []*nodeWithStyles {
	nodeList := make([]*nodeWithStyles, 0, 100)
	pStyle := createArtificialNodes(0, []*cdp.Node{})

	nodeList = append(nodeList, pStyle)

	needsChildrenPos := 0

	for i := 1; i < 100; i++ {
		childSlice := make([]*cdp.Node, 0, 2)
		artNode := createArtificialNodes(int64(i), childSlice)
		nodeList = append(nodeList, artNode)
		nodeList[needsChildrenPos].node.Children = append(nodeList[needsChildrenPos].node.Children, artNode.node)

		if len(nodeList[needsChildrenPos].node.Children) == 2 {
			needsChildrenPos += 1
		}
	}

	return nodeList
}

func TestParseThroughNodes(t *testing.T) {
	nodeTree := buildNodeTree()
	nodeList := parseThroughNodes(nodeTree)

	if len(nodeList) != 100 {
		t.Fatal("did nto parse nodes with proper length")
	}
}

func TestNavigateValidate(t *testing.T) {
	u := navigateWebPage{
		url: "https://bench-ai.com",
	}

	u1 := navigateWebPage{
		url: "http://bench-ai.com",
	}

	// u2 is a illegal var name

	u3 := navigateWebPage{
		url: "ht//bench-ai.com",
	}

	if u.validate() != nil || u1.validate() != nil {
		t.Fatal("function validate for navigateWebPage failed to validate passing urls")
	}

	if u3.validate() == nil {
		t.Fatal("function validate for navigateWebPage failed to detect failing url")
	}
}

func TestNavigateInitFromJson(t *testing.T) {
	data := `{"url": "https://bench-ai.com"}`
	u := navInitFromJson([]byte(data))

	if u.url != "https://bench-ai.com" {
		t.Fatal("function navInitFromJson failed to convert bytes to struct")
	}
}

func tempDir() (error, string) {
	td, err := os.MkdirTemp("", "test_temp_func")
	return err, td
}

func TestFpsValidate(t *testing.T) {
	fps := fullPageScreenShot{
		savePath: "path/to/save.jpg",
		quality:  10,
	}

	if fps.validate() != nil {
		t.Fatal("fullPageScreenShot struct is valid")
	}

	fps.savePath = "path/to/save.txt"

	if fps.validate() == nil {
		t.Fatal("fullPageScreenShot struct cannot detect invalid savePath")
	}

	fps.savePath = "path/to/save.jpg"
	fps.quality = 0

	if fps.validate() == nil {
		t.Fatal("fullPageScreenShot struct cannot detect quality of 0")
	}
}

func TestFpsInitFromJson(t *testing.T) {
	data := `{"quality": 10, "snapshot_name": "s1", "name": "test.jpg"}`

	byteSlice := []byte(data)

	err, pth := tempDir()

	if err != nil {
		t.Fatal(err)
	}

	pFps := fpsInitFromJson(byteSlice, pth)

	if pFps.quality != 10 {
		t.Fatal("did not properly unmarshall struct field quality")
	}

	if !strings.Contains(pFps.savePath, "s1") {
		t.Fatal("savePath does not contain snapshot name")
	}

	if !strings.Contains(pFps.savePath, pth) {
		t.Fatal("savePath does not contain session path")
	}

	if !strings.Contains(pFps.savePath, "test.jpg") {
		t.Fatal("savePath does not contain file name")
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(pth)
}

func TestValidateElementScreenshot(t *testing.T) {
	es := elementScreenshot{
		savePath: "path/to/my/screeshot.jpg",
		scale:    10,
		selector: "xpath/asdsdasd",
	}

	if es.validate() != nil {
		t.Fatal("valid elementScreenshot was rejected")
	}

	es.scale = 0
	if es.validate() == nil {
		t.Fatal("invalid elementScreenshot scale of 0 was accepted")
	}

	es.scale = -1
	if es.validate() == nil {
		t.Fatal("invalid elementScreenshot scale of < 0 was accepted")
	}

	es.scale = 10
	es.savePath = "path/to/my/file"
	if es.validate() == nil {
		t.Fatal("invalid elementScreenshot savePath was accepted")
	}

	es.savePath = "path/to/my/screeshot.jpg"
	es.selector = ""
	if es.validate() == nil {
		t.Fatal("empty xpath was accepted")
	}
}

func TestElemInitFromJson(t *testing.T) {
	data := `{"scale": 10, "snapshot_name": "s1", "name": "test.jpg", "selector": "xpath/123/tada"}`
	dBytes := []byte(data)

	err, pth := tempDir()

	if err != nil {
		t.Fatal(err)
	}

	el := elemInitFromJson(dBytes, pth)

	if el.scale != 10 {
		t.Fatal("could not marshall scale")
	}

	if el.selector != "xpath/123/tada" {
		t.Fatal("could not marshall selector")
	}

	if !strings.Contains(el.savePath, "s1") ||
		!strings.Contains(el.savePath, "test.jpg") ||
		!strings.Contains(el.savePath, pth) {
		t.Fatal("could not marshall selector")
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatal(err)
		}
	}(pth)
}

func TestClickValidate(t *testing.T) {
	c := click{
		selector: "body",
	}

	if c.validate() != nil {
		t.Fatal("valid click was rejected")
	}

	c.selector = ""
	if c.validate() == nil {
		t.Fatal("empty click selector was accepted")
	}
}

func TestClickInitFromJson(t *testing.T) {
	data := `{"selector": "data", "query_type": "search"}`
	data1 := `{"selector": "data", "query_type": "searc"}`

	if _, err := clickInitFromJson([]byte(data1)); err == nil {
		t.Fatal(err)
	}

	if _, err := clickInitFromJson([]byte(data)); err != nil {
		t.Fatal(err)
	}
}

func TestSleepInitFromJson(t *testing.T) {
	data := `{"ms": 100000}`

	d := sleepInitFromJson([]byte(data))

	if d.ms != 100000 {
		t.Fatal("failed to marshall sleep time")
	}
}

func TestHtmlInitFromJson(t *testing.T) {
	data := `{"snapshot_name": "s1", "selector": "body"}`
	dBytes := []byte(data)

	err, pth := tempDir()
	if err != nil {
		t.Fatal(err)
	}

	_, err = htmlInitFromJson(dBytes, pth)
	if err != nil {
		t.Fatal("could not create valid html struct")
	}

	if os.RemoveAll(pth) != nil {
		t.Fatal(err)
	}

	data = `{"snapshot_name": "s1"}`
	dBytes = []byte(data)

	err, pth = tempDir()
	if err != nil {
		t.Fatal(err)
	}

	h, err := htmlInitFromJson(dBytes, pth)
	if err != nil {
		t.Fatal("valid html command rejected")
	}

	if h.selector != "html" {
		t.Fatal("selector did not default to html")
	}

	if os.RemoveAll(pth) != nil {
		t.Fatal(err)
	}

	data = `{"snapshot_name": "", "selector": "s1"}`
	dBytes = []byte(data)

	err, pth = tempDir()
	if err != nil {
		t.Fatal(err)
	}

	_, err = htmlInitFromJson(dBytes, pth)
	if err == nil {
		t.Fatal("accepted invalid snapshot")
	}

	if os.RemoveAll(pth) != nil {
		t.Fatal(err)
	}

	defer func() {
		if os.RemoveAll(pth) != nil {
			t.Fatal(err)
		}
	}()
}

func TestValidateHtml(t *testing.T) {
	h := html{
		selector: "",
	}

	if h.validate() == nil {
		t.Fatal("accepted empty selector")
	}

	h.selector = "body"

	if h.validate() != nil {
		t.Fatal("rejected valid html command")
	}
}

func TestValidateNodeCollect(t *testing.T) {
	nc := nodeCollect{
		selector: "",
	}

	if nc.validate() == nil {
		t.Fatal("empty selector was not rejected")
	}

	nc.selector = "body"

	if nc.validate() != nil {
		t.Fatal("valid node collect struct rejected")
	}
}
