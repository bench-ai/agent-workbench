package browser

import (
	"agent/helper"
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/chromedp"
	"image"
	"image/jpeg"
	"log"
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

// writeImg
/*
saves the current image for writing
*/
func writeImg(
	snapshot string,
	imageQuality uint8,
	ctx context.Context) (error, *imageMetaData) {

	var byteSlice []byte

	imgMetaData := imageMetaData{
		snapShotName: snapshot,
		imageName:    "page_image.png",
		byteData:     &byteSlice,
	}

	err := chromedp.FullScreenshot(&byteSlice, int(imageQuality)).Do(ctx)
	if err != nil {
		return err, nil
	}

	return nil, &imgMetaData
}

// writeHtml
/*
saves the current html for writing
*/
func writeHtml(ctx context.Context) (error, string) {
	var html string
	err := chromedp.OuterHTML("body", &html).Do(ctx)
	if err != nil {
		return err, ""
	}

	return nil, html
}

// saveSnapshot
/*
saves a snapshot with all the wanted files
*/
func saveSnapshot(
	htmlMap map[string]*string,
	saveNode map[string]*[]*nodeWithStyles,
	nodeSlice []*nodeWithStyles,
	fullPageImgSlice *[]*imageMetaData,
	imgMD *imageMetaData,
	ctx context.Context,
	snapshot string) error {

	if fullPageImgSlice != nil {
		*fullPageImgSlice = append(*fullPageImgSlice, imgMD)
	}

	if htmlMap != nil {
		err, html := writeHtml(ctx)
		if err != nil {
			return err
		}
		htmlMap[snapshot] = &html
	}

	if saveNode != nil {
		saveNode[snapshot] = &nodeSlice
	}

	return nil
}

/**
TODO In order of priority
2) add code to click and screenshot all elements
3) add check limit
3) unit test
5) speed up the iterator action
*/

// htmlIteratorAction
/*
collects all unique transition snapshots
*/
func htmlIteratorAction(
	iterLimit uint16,
	pauseTime uint32,
	startingSnapshot uint8,
	snapshotName string,
	imageQuality uint8,
	htmlMap map[string]*string,
	saveNode map[string]*[]*nodeWithStyles,
	fullPageImgSlice *[]*imageMetaData,
) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.ActionFunc(func(c context.Context) error {

			startTime := time.Now()
			var pByteCollection []*[]byte
			nodeSlice := make([]*nodeWithStyles, 0, 10)

			err := chromedp.Sleep(time.Second * 5).Do(c) //wait for website to settle down
			if err != nil {
				return err
			}

			// save initial snapshot
			currentTime := time.Now()
			diff := currentTime.Sub(startTime)
			snapshot := fmt.Sprintf(
				"%s_%d_%d_ms", snapshotName, startingSnapshot, diff.Milliseconds(),
			)
			err, imgMD := writeImg(snapshot, imageQuality, c)
			pByteCollection = append(pByteCollection, imgMD.byteData)
			if err != nil {
				return err
			}
			err = populatedNodeAction("body", true, true, &nodeSlice).Do(c)
			if err != nil {
				return err
			}
			err = saveSnapshot(
				htmlMap,
				saveNode,
				nodeSlice,
				fullPageImgSlice,
				imgMD,
				c,
				snapshot)
			if err != nil {
				return err
			}

			nodeMap := nodeToMap(nodeSlice)

			ticker := time.NewTicker(time.Duration(pauseTime) * time.Millisecond)

			hitLimit := 10 //program will only return after consecutively not having a unique transition this many times
			hitCount := 0
			count := uint16(0)

			for range ticker.C {
				count++
				startingSnapshot++

				currentTime = time.Now()
				diff = currentTime.Sub(startTime)
				snapshot = fmt.Sprintf(
					"%s_%d_%d_ms", snapshotName, startingSnapshot, diff.Milliseconds(),
				)
				err, imgMD = writeImg(snapshot, imageQuality, c)
				if err != nil {
					return err
				}
				pByteCollection = append(pByteCollection, imgMD.byteData)
				currentNodeSlice := make([]*nodeWithStyles, 0, 10)
				err = populatedNodeAction("body", true, true, &currentNodeSlice).Do(c)
				if err != nil {
					return err
				}

				currentNodeMap := nodeToMap(currentNodeSlice)
				nodeMap = mergeNodeMap(currentNodeMap, nodeMap)

				if err != nil {
					return err
				}

				// check whether the page has transitioned
				if checkPageTransition(nodeMap, pByteCollection) {

					hitCount = 0

					err = saveSnapshot(
						htmlMap,
						saveNode,
						currentNodeSlice,
						fullPageImgSlice,
						imgMD,
						c,
						snapshot)

					if err != nil {
						return err
					}

				} else {
					// if page has not transitioned exit if hits has been hit
					hitCount++
					if hitCount == hitLimit {
						return nil
					}
				}

				if count == iterLimit {
					return nil
				}
			}

			return nil
		}),
	}
}
