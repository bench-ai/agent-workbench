package chrome

import (
	"agent/helper"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/chromedp"
	"image"
	"image/jpeg"
	"log"
	"path/filepath"
	"time"
)

// compareImages
/**
checks whether images (which are jpegs are equal)

TODO: Add potential percentage similarity i.e. return true > img1 matches 90% of img2
*/
func compareImages(imgOne, imgTwo *image.Image) bool {

	boundsOne := (*imgOne).Bounds()
	boundsTwo := (*imgTwo).Bounds()

	if boundsOne.Dx() != boundsTwo.Dx() || boundsOne.Dy() != boundsTwo.Dy() {
		return false
	}

	for y := boundsOne.Min.Y; y < boundsOne.Max.Y; y++ {
		for x := boundsOne.Min.X; x < boundsOne.Max.X; x++ {
			pixelOne := (*imgOne).At(x, y)
			pixelTwo := (*imgTwo).At(x, y)
			if pixelOne != pixelTwo {
				return false
			}
		}
	}

	return true
}

// containsImage
/*
check if an image is present in array of images
*/
func containsImage(newBytes *[]byte, imgSlice []*[]byte) bool {

	newImage, err := jpeg.Decode(bytes.NewReader(*newBytes))

	if err != nil {
		log.Fatal("could not convert image to jpeg", err)
	}

	for _, oldBytes := range imgSlice {

		oldImage, _, err := image.Decode(bytes.NewReader(*oldBytes))

		if err != nil {
			log.Fatal("could not convert image to jpeg", err)
		}

		if compareImages(&oldImage, &newImage) {
			return true
		}
	}

	return false
}

// nodesAreEqual
/*
check with two nodes share the same metadata (in some categories)
*/
func nodesAreEqual(n1, n2 *cdp.Node) bool {
	if n1.NodeID != n2.NodeID {
		return false
	}

	if n1.ParentID != n2.ParentID {
		return false
	}

	if n1.BackendNodeID != n2.BackendNodeID {
		return false
	}

	if n1.NodeType != n2.NodeType {
		return false
	}

	if n1.NodeName != n2.NodeName {
		return false
	}

	if n1.LocalName != n2.LocalName {
		return false
	}

	if n1.NodeValue != n2.NodeValue {
		return false
	}

	if len(n2.Attributes) != len(n1.Attributes) {
		return false
	}

	for _, i := range n2.Attributes {
		if !helper.Contains[string](n1.Attributes, i) {
			return false
		}
	}

	return true
}

// flattenNode
/*
traverses the node tree using BFS, and adds all child nodes to the node pointer slice
*/
func flattenNode(nodeSlice []*cdp.Node) []*cdp.Node {
	deleteByIndex := helper.DeleteByIndex[*cdp.Node]

	var newNodeSlice []*cdp.Node

	for len(nodeSlice) > 0 {
		nodeSlice = append(nodeSlice, nodeSlice[0].Children...)
		newNodeSlice = append(newNodeSlice, nodeSlice[0])
		_, nodeSlice = deleteByIndex(nodeSlice, 0)
	}

	return newNodeSlice
}

// equalStyles
/*
checks whether 2 cssStyles are exactly the same
*/
func equalStyles(originalStyles, newStyles []*css.ComputedStyleProperty) bool {

	if len(originalStyles) != len(newStyles) {
		return false
	}

	for i, prop := range newStyles {
		if prop.Name != originalStyles[i].Name || prop.Value != originalStyles[i].Value {
			return false
		}
	}

	return true
}

// nodeToMap
/*
creates a map of the nodeId with a value of the array of all node instances with the same ID
useful for comparisons to see if node repeats
*/
func nodeToMap(nodeSLice []*nodeWithStyles) map[cdp.NodeID][]*nodeWithStyles {
	nodeMap := map[cdp.NodeID][]*nodeWithStyles{}

	for _, styledNode := range nodeSLice {
		nodeMap[styledNode.node.NodeID] = []*nodeWithStyles{
			styledNode,
		}
	}

	return nodeMap
}

// mergeNodeMap
/*
merges 2 node maps
*/
func mergeNodeMap(newMap, originalMap map[cdp.NodeID][]*nodeWithStyles) map[cdp.NodeID][]*nodeWithStyles {
	for key, val := range newMap {
		if slice, ok := originalMap[key]; ok {
			originalMap[key] = append(slice, val...)
		} else {
			originalMap[key] = val
		}

	}

	return originalMap
}

// checkPageTransition
/*
Checks if the webpage has changed to a unique look. Each transition is compared against previous transitions
*/
func checkPageTransition(
	dataMap map[cdp.NodeID][]*nodeWithStyles,
	byteCollection []*[]byte,
) bool {
	lastByte := byteCollection[len(byteCollection)-1]

	// checks whether the current image snapshot is present in all other snapshots if so return false
	if containsImage(lastByte, byteCollection[:len(byteCollection)-1]) {
		return false
	}

	// if the image is different it most likely means the page has
	// transitioned. However, we double-check using the nodes

	// iterate through each node in a webpage
	for _, nodeList := range dataMap {
		lastNode := nodeList[len(nodeList)-1] // represents the current node for the most recent snapshot
		unique := true

		// iterate through all other node snapshots
		for _, node := range nodeList[:len(nodeList)-1] {
			if nodesAreEqual(node.node, lastNode.node) && equalStyles(node.cssStyles, lastNode.cssStyles) {
				unique = false
				break
			}
		}

		if unique {
			// found no other nodes matching this one in the snapshot return true
			return unique
		}
	}

	return false
}

// getImg
/*
saves the current image for writing
*/
func getImg(
	imageQuality uint8,
	ctx context.Context) (error, []byte) {

	var buffer []byte

	err := chromedp.Tasks{
		takeFullPageScreenshot(imageQuality, &buffer),
	}.Do(ctx)

	if err != nil {
		return err, nil
	}

	return err, buffer
}

func writeImg(
	savePath string,
	job *fileJob,
	buffer []byte) {

	filePath := filepath.Join(savePath, "images", "fullPage.jpg")
	job.writeBytes(buffer, filePath)
}

// getHtml
/*
saves the current html for writing
*/
func getHtml(
	ctx context.Context,
) (error, string) {

	var text string
	err := collectHtml("body", &text).Do(ctx)

	if err != nil {
		return err, ""
	}

	return err, text
}

func writeHtml(
	savePath,
	text string,
	job *fileJob) {
	filePath := filepath.Join(savePath, "html.txt")
	job.writeBytes([]byte(text), filePath)
}

// getNodes
/*
saves a snapshot with all the wanted files
*/
func getNodes(
	ctx context.Context) (error, []*nodeWithStyles) {

	var nodeSlice []*nodeWithStyles

	err := populatedNode(
		"body",
		true,
		true,
		true,
		&nodeSlice).Do(ctx)

	if err != nil {
		log.Fatal("could not convert nodes to json bytes")
	}

	return err, nodeSlice
}

func writeNodes(
	nodeSlice []*nodeWithStyles,
	savePath string,
	job *fileJob) {

	filePath := filepath.Join(savePath, "nodes.json")

	byteSlice, err := json.Marshal(parseThroughNodes(nodeSlice))

	if err != nil {
		log.Fatal("could not convert nodes to json bytes")
	}

	job.writeBytes(byteSlice, filePath)
}

type htmlIterator struct {
	iterLimit        uint16
	restTimeMs       uint32
	startingSnapshot uint8
	snapshotName     string
	imageQuality     uint8
	sessionPath      string
	saveHtml         bool
	saveNode         bool
	saveImage        bool
}

func htmlIterInitFromJson(jsonBytes []byte, sessionPath string) *htmlIterator {
	type body struct {
		IterLimit         *uint16 `json:"iter_limit"`
		PauseTime         *uint32 `json:"pause_time"`
		StartingSnapshot  *uint8  `json:"starting_snapshot"`
		SnapshotName      string  `json:"snapshot_name"`
		SaveHtml          bool    `json:"save_html"`
		SaveNode          bool    `json:"save_node"`
		SaveFullPageImage bool    `json:"save_full_page_image"`
		ImageQuality      *uint8  `json:"image_quality"`
	}

	bdy := body{}
	err := json.Unmarshal(jsonBytes, &bdy)

	if err != nil {
		log.Fatal(err)
	}

	iter := htmlIterator{
		snapshotName: bdy.SnapshotName,
		saveImage:    bdy.SaveFullPageImage,
		saveHtml:     bdy.SaveHtml,
		saveNode:     bdy.SaveNode,
	}

	if bdy.IterLimit == nil {
		iter.iterLimit = 10
	} else {
		iter.iterLimit = *bdy.IterLimit
	}

	if bdy.PauseTime == nil {
		iter.restTimeMs = 1000
	} else {
		iter.restTimeMs = *bdy.PauseTime
	}

	if bdy.StartingSnapshot == nil {
		iter.startingSnapshot = 0
	} else {
		iter.startingSnapshot = *bdy.StartingSnapshot
	}

	if bdy.ImageQuality == nil {
		iter.imageQuality = 10
	} else {
		iter.imageQuality = *bdy.ImageQuality
	}

	iter.sessionPath = sessionPath

	return &iter
}

func (h *htmlIterator) validate() error {
	if h.snapshotName == "" {
		log.Fatal("snapshot name for iter html cannot be blank")
	}

	return nil
}

// htmlIteratorAction
/*
collects all unique transition snapshots
*/
func (h *htmlIterator) getAction(job *fileJob) chromedp.ActionFunc {
	return func(c context.Context) error {

		startTime := time.Now()
		var pByteCollection []*[]byte
		//nodeSlice := make([]*nodeWithStyles, 0, 10)

		err := chromedp.Sleep(time.Second * 5).Do(c) //wait for website to settle down
		if err != nil {
			return err
		}

		// save initial snapshot
		currentTime := time.Now()
		diff := currentTime.Sub(startTime)
		snapshot := fmt.Sprintf(
			"%s_%d_%d_ms", h.snapshotName, h.startingSnapshot, diff.Milliseconds(),
		)

		writePath := createSnapshotFolder(h.sessionPath, snapshot)
		err, imgByte := getImg(h.imageQuality, c)

		if err == nil {
			if h.saveImage {
				writeImg(writePath, job, imgByte)
			}
		} else {
			return err
		}

		err, nodeSlice := getNodes(c)
		if err == nil {
			if h.saveNode {
				writeNodes(nodeSlice, writePath, job)
			}
		} else {
			return err
		}

		if h.saveHtml {
			if err, text := getHtml(c); err != nil {
				return err
			} else {
				writeHtml(writePath, text, job)
			}
		}

		pByteCollection = append(pByteCollection, &imgByte)

		nodeMap := nodeToMap(nodeSlice)

		ticker := time.NewTicker(time.Duration(h.restTimeMs) * time.Millisecond)

		hitLimit := 10 //program will only return after consecutively not having a unique transition this many times
		hitCount := 0
		count := uint16(0)

		for range ticker.C {
			count++
			h.startingSnapshot++

			currentTime = time.Now()
			diff = currentTime.Sub(startTime)
			snapshot = fmt.Sprintf(
				"%s_%d_%d_ms", h.snapshotName, h.startingSnapshot, diff.Milliseconds(),
			)

			err, imgBytes := getImg(h.imageQuality, c)
			if err != nil {
				return err
			}

			pByteCollection = append(pByteCollection, &imgBytes)

			err, currentNodeStyles := getNodes(c)
			if err != nil {
				return err
			}

			currentNodeMap := nodeToMap(currentNodeStyles)
			nodeMap = mergeNodeMap(currentNodeMap, nodeMap)

			if err != nil {
				return err
			}

			// check whether the page has transitioned
			if checkPageTransition(nodeMap, pByteCollection) {
				writePath = createSnapshotFolder(h.sessionPath, snapshot)

				hitCount = 0

				if h.saveNode {
					writeNodes(currentNodeStyles, writePath, job)
				}

				if h.saveImage {
					writeImg(writePath, job, imgBytes)
				}

				if h.saveHtml {
					if err, text := getHtml(c); err == nil {
						writeHtml(writePath, text, job)
					} else {
						return err
					}
				}

			} else {
				// if page has not transitioned exit if hits has been hit
				hitCount++
				if hitCount == hitLimit {
					return nil
				}
			}

			if count == h.iterLimit {
				return nil
			}
		}

		return nil
	}
}
