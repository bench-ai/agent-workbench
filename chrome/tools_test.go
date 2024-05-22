package chrome

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"math/rand"
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
