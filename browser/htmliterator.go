package browser

import (
	"agent/helper"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/chromedp"
	"time"
)

func compareNodes(n1, n2 *cdp.Node) bool {
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

	for _, i := range n2.Attributes {
		if !helper.Contains[string](n1.Attributes, i) {
			return false
		}
	}

	return true
}

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

func nodeToMap(nodeSLice []*cdp.Node, ctx context.Context) (error, map[cdp.NodeID][]*nodeWithStyles) {
	nodeMap := map[cdp.NodeID][]*nodeWithStyles{}

	for _, node := range nodeSLice {

		styles, err := css.GetComputedStyleForNode(node.NodeID).Do(ctx)

		if err != nil {
			return err, nil
		}

		styledNode := nodeWithStyles{
			node:      node,
			cssStyles: styles,
		}

		nodeMap[node.NodeID] = []*nodeWithStyles{
			&styledNode,
		}
	}

	return nil, nodeMap
}

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

func checkPageTransition(dataMap map[cdp.NodeID][]*nodeWithStyles) bool {
	for _, nodeList := range dataMap {
		lastNode := nodeList[len(nodeList)-1]

		unique := true
		for _, node := range nodeList[:len(nodeList)-1] {
			if compareNodes(node.node, lastNode.node) && equalStyles(node.cssStyles, lastNode.cssStyles) {
				unique = false
			}
		}

		if unique {
			return unique
		}
	}

	return false
}

func writeImg(snapshot string, imageQuality uint8, ctx context.Context) (error, *imageMetaData) {
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

func writeHtml(ctx context.Context) (error, string) {
	var html string
	err := chromedp.OuterHTML("body", &html).Do(ctx)
	if err != nil {
		return err, ""
	}

	return nil, html
}

func saveSnapshot(
	startingSnapshot uint8,
	snapshotName string,
	imageQuality uint8,
	htmlMap map[string]*string,
	saveNode *[]*cdp.Node,
	fullPageImgSlice *[]*imageMetaData,
	ctx context.Context,
	startTime time.Time) error {

	currentTime := time.Now()

	diff := currentTime.Sub(startTime)

	snapshot := fmt.Sprintf(
		"%s_%d_%d_ms", snapshotName, startingSnapshot, diff.Milliseconds(),
	)

	if fullPageImgSlice != nil {
		err, imgMD := writeImg(snapshot, imageQuality, ctx)
		if err != nil {
			return err
		}
		*fullPageImgSlice = append(*fullPageImgSlice, imgMD)
	}

	if htmlMap != nil {
		err, html := writeHtml(ctx)
		if err != nil {
			return err
		}
		htmlMap[snapshot] = &html
	}

	return nil
}

/**
TODO In order of priority
1) Add ability to write nodes
2) add code to click and screenshot all elements
3) unit test
4) add ability to save css
5) speed up the iterator action
*/

func htmlIteratorAction(
	iterLimit uint16,
	pauseTime uint32,
	startingSnapshot uint8,
	snapshotName string,
	imageQuality uint8,
	htmlMap map[string]*string,
	saveNode *[]*cdp.Node,
	fullPageImgSlice *[]*imageMetaData,
) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.ActionFunc(func(c context.Context) error {

			startTime := time.Now()

			var nodeSlice []*cdp.Node
			err := chromedp.Sleep(time.Second * 5).Do(c)
			if err != nil {
				return err
			}

			err = getPopulatedNodes("body", &nodeSlice).Do(c)
			if err != nil {
				return err
			}

			count := uint16(0)

			err = saveSnapshot(
				startingSnapshot,
				snapshotName,
				imageQuality,
				htmlMap,
				saveNode,
				fullPageImgSlice,
				c,
				startTime)

			if err != nil {
				return err
			}

			err, nodeMap := nodeToMap(flattenNode(nodeSlice), c)

			if err != nil {
				return err
			}

			ticker := time.NewTicker(time.Duration(pauseTime) * time.Millisecond)

			for range ticker.C {
				count++
				startingSnapshot++

				var currentNodeSlice []*cdp.Node
				err = getPopulatedNodes("body", &currentNodeSlice).Do(c)
				if err != nil {
					return err
				}

				err, currentNodeMap := nodeToMap(flattenNode(nodeSlice), c)
				if err != nil {
					return err
				}

				nodeMap = mergeNodeMap(nodeMap, currentNodeMap)

				if checkPageTransition(nodeMap) {
					err = saveSnapshot(
						startingSnapshot,
						snapshotName,
						imageQuality,
						htmlMap,
						saveNode,
						fullPageImgSlice,
						c,
						startTime)

					if err != nil {
						return err
					}
				} else {
					break
				}

				if count == iterLimit {
					return nil
				}
			}

			return nil
		}),
	}
}
