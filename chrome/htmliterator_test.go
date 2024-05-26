package chrome

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"image"
	"image/color"
	"math/rand"
	"os"
	"testing"
)

func newRandomImage(x, y, x1, y1 int) image.Image {
	rect1 := image.Rect(x, y, x1, y1)
	randImg := image.NewRGBA(rect1)

	bounds := randImg.Bounds()

	randInt := func() uint8 {
		return uint8(rand.Intn(255))
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBA{R: randInt(), G: randInt(), B: randInt(), A: 255}
			randImg.Set(x, y, rgba)
		}
	}

	return randImg
}

func TestCompareImages(t *testing.T) {
	r1 := newRandomImage(0, 0, 100, 100)
	r2 := newRandomImage(0, 0, 100, 100)
	r3 := newRandomImage(0, 0, 95, 95)

	if !compareImages(&r1, &r1) {
		t.Error("identical images are not recognized as the same")
	}

	if compareImages(&r1, &r2) {
		t.Error("different images are recognized as identical")
	}

	if compareImages(&r2, &r3) {
		t.Error("different sized images are recognized as identical")
	}
}

func TestNodesAreEqual(t *testing.T) {
	n1 := cdp.Node{
		NodeID:        cdp.NodeID(1),
		ParentID:      cdp.NodeID(0),
		BackendNodeID: cdp.BackendNodeID(1),
		NodeType:      cdp.NodeTypeElement,
		NodeName:      "test",
		LocalName:     "localTest",
		NodeValue:     "1",
		Attributes: []string{
			"id", "my-node", "color", "brown",
		},
	}

	n2 := cdp.Node{
		NodeID:        cdp.NodeID(2),
		ParentID:      cdp.NodeID(1),
		BackendNodeID: cdp.BackendNodeID(11),
		NodeType:      cdp.NodeTypeText,
		NodeName:      "test1",
		LocalName:     "local1Test",
		NodeValue:     "1",
		Attributes: []string{
			"id", "my-node", "brown",
		},
	}

	n3 := cdp.Node{
		NodeID:        cdp.NodeID(1),
		ParentID:      cdp.NodeID(0),
		BackendNodeID: cdp.BackendNodeID(1),
		NodeType:      cdp.NodeTypeElement,
		NodeName:      "test",
		LocalName:     "localTest",
		NodeValue:     "1",
		Attributes: []string{
			"id", "my-node", "color",
		},
	}

	if !nodesAreEqual(&n1, &n1) {
		t.Error("unable to recognize equal nodes")
	}

	if nodesAreEqual(&n1, &n2) {
		t.Error("recognized unequal nodes as equal")
	}

	if nodesAreEqual(&n1, &n3) {
		t.Error("recognized unequal nodes as equal")
	}
}

func TestFlattenNode(t *testing.T) {
	n1 := cdp.Node{
		NodeID:        cdp.NodeID(1),
		ParentID:      cdp.NodeID(0),
		BackendNodeID: cdp.BackendNodeID(1),
		NodeType:      cdp.NodeTypeElement,
		NodeName:      "test",
		LocalName:     "localTest",
		NodeValue:     "1",
		Attributes: []string{
			"id", "my-node", "color", "brown",
		},
	}

	n2 := cdp.Node{
		NodeID:        cdp.NodeID(1),
		ParentID:      cdp.NodeID(0),
		BackendNodeID: cdp.BackendNodeID(2),
		NodeType:      cdp.NodeTypeText,
		NodeName:      "test1",
		LocalName:     "local1Test",
		NodeValue:     "1",
		Attributes: []string{
			"id", "my-node", "brown",
		},
		Children: []*cdp.Node{
			&n1,
		},
	}

	n3 := cdp.Node{
		NodeID:        cdp.NodeID(3),
		ParentID:      cdp.NodeID(1),
		BackendNodeID: cdp.BackendNodeID(2),
		NodeType:      cdp.NodeTypeElement,
		NodeName:      "test",
		LocalName:     "localTest",
		NodeValue:     "1",
		Attributes: []string{
			"id", "my-node", "color",
		},
		Children: []*cdp.Node{
			&n1, &n2,
		},
	}

	if len(flattenNode([]*cdp.Node{&n3})) == 5 {
		t.Error("flatten length did not match estimated length")
	}
}

func TestEqualStyles(t *testing.T) {
	styleList := []*css.ComputedStyleProperty{
		{
			Name:  "name",
			Value: "data",
		},
		{
			Name:  "name",
			Value: "data",
		},
	}

	styleList1 := []*css.ComputedStyleProperty{
		{
			Name:  "name1",
			Value: "data1",
		},
	}

	styleList2 := []*css.ComputedStyleProperty{
		{
			Name:  "name1",
			Value: "data1",
		},
		{
			Name:  "name",
			Value: "data",
		},
	}

	if !equalStyles(styleList, styleList) {
		t.Error("did not recognize equal styles")
	}

	if equalStyles(styleList1, styleList) {
		t.Error("did not recognize different styles")
	}

	if equalStyles(styleList, styleList2) {
		t.Error("did not recognize different styles")
	}
}

func TestNodeToMap(t *testing.T) {
	n1 := &nodeWithStyles{
		cssStyles: []*css.ComputedStyleProperty{
			{
				Name:  "name",
				Value: "data",
			},
			{
				Name:  "name",
				Value: "data",
			},
		},
		node: &cdp.Node{
			NodeID:        cdp.NodeID(1),
			ParentID:      cdp.NodeID(0),
			BackendNodeID: cdp.BackendNodeID(1),
			NodeType:      cdp.NodeTypeElement,
			NodeName:      "test",
			LocalName:     "localTest",
			NodeValue:     "1",
			Attributes: []string{
				"id", "my-node", "color", "brown",
			},
		},
	}

	n2 := &nodeWithStyles{
		cssStyles: []*css.ComputedStyleProperty{
			{
				Name:  "name",
				Value: "data1",
			},
			{
				Name:  "name",
				Value: "data2",
			},
		},
		node: &cdp.Node{
			NodeID:        cdp.NodeID(3),
			ParentID:      cdp.NodeID(0),
			BackendNodeID: cdp.BackendNodeID(1),
			NodeType:      cdp.NodeTypeElement,
			NodeName:      "test",
			LocalName:     "localTest",
			NodeValue:     "1",
			Attributes: []string{
				"id", "my-node", "color", "brown",
			},
		},
	}

	n3 := &nodeWithStyles{
		cssStyles: []*css.ComputedStyleProperty{
			{
				Name:  "name",
				Value: "data1",
			},
			{
				Name:  "name",
				Value: "data2",
			},
		},
		node: &cdp.Node{
			NodeID:        cdp.NodeID(4),
			ParentID:      cdp.NodeID(0),
			BackendNodeID: cdp.BackendNodeID(1),
			NodeType:      cdp.NodeTypeElement,
			NodeName:      "test",
			LocalName:     "localTest",
			NodeValue:     "1",
			Attributes: []string{
				"id", "my-node", "color", "brown",
			},
		},
	}

	n4 := &nodeWithStyles{
		cssStyles: []*css.ComputedStyleProperty{
			{
				Name:  "name",
				Value: "data1",
			},
			{
				Name:  "name",
				Value: "data2",
			},
		},
		node: &cdp.Node{
			NodeID:        cdp.NodeID(1),
			ParentID:      cdp.NodeID(0),
			BackendNodeID: cdp.BackendNodeID(1),
			NodeType:      cdp.NodeTypeElement,
			NodeName:      "test",
			LocalName:     "localTest",
			NodeValue:     "1",
			Attributes: []string{
				"id", "my-node", "color", "brown",
			},
		},
	}

	s1 := []*nodeWithStyles{
		n1, n2,
	}

	s2 := []*nodeWithStyles{
		n3,
	}

	s3 := []*nodeWithStyles{
		n4,
	}

	m1 := nodeToMap(s1)
	m2 := nodeToMap(s2)
	m3 := nodeToMap(s3)
	m4 := mergeNodeMap(m1, m2)
	m5 := mergeNodeMap(m1, m3)

	if len(m4) != 3 {
		t.Error("found inaccurate size in map")
	}

	if len(m5) != 2 {
		t.Error("found inaccurate size in map")
	}

	if len(m5[cdp.NodeID(1)]) != 2 {
		t.Error("found inaccurate size in map slice len")
	}
}

func TestHtmlIterInitFromJson(t *testing.T) {
	jBytes := []byte("{}")

	err, sess := tempDir()

	if err != nil {
		t.Fatal(err)
	}

	def := htmlIterInitFromJson(jBytes, sess)

	if def.iterLimit == 0 {
		t.Fatalf("iter limit was not assigned a proper default")
	}

	if def.restTimeMs == 0 {
		t.Fatalf("rest time was not assigned a proper default")
	}

	if def.imageQuality == 0 {
		t.Fatalf("image quality was not assigned a proper default")
	}

	defer func() {
		if err := os.RemoveAll(sess); err != nil {
			t.Fatal(err)
		}
	}()
}

func TestHtmlValidate(t *testing.T) {
	jBytes := []byte(`{"snapshot_name": "s1"}`)

	err, sess := tempDir()

	if err != nil {
		t.Fatal(err)
	}

	def := htmlIterInitFromJson(jBytes, sess)

	if def.validate() != nil {
		t.Fatal("valid snapshot name was rejected")
	}

	def.snapshotName = ""

	if def.validate() == nil {
		t.Fatal("invalid snapshot name was accepted")
	}
}
