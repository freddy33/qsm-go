package m3space

import "fmt"

const (
	noDimmer              = 1.0
	defaultGreyDimmer     = 0.7
	defaultOldEventDimmer = 0.3
)

type ObjectType uint8

const (
	AxeX        ObjectType = iota
	AxeY
	AxeZ
	NodeEmpty
	NodeActive
	Connection00
)

type SpaceDrawingElement interface {
	// Key of the drawing element to point to the OpenGL buffer to render
	Key() ObjectType
	// The translation point to apply to the OpenGL model, since all the above are drawn at the origin
	Pos() *Point
	// Return the obj_color int for the shader program
	Color(blinkValue float64) int32
	// Return the obj_dimmer int for the shader program
	Dimmer(blinkValue float64) float32
	// Display flag
	Display() bool
}

type SpaceDrawingFilter struct {
	// Display grey empty nodes or not
	DisplayEmptyNodes bool
	// Display grey empty connections or not
	DisplayEmptyConnections bool
	// Distance from latest to display event outgrowth
	EventOutgrowthThreshold Distance
	// The events of certain colors to display. This is a mask.
	EventColorMask uint8
	// The outgrowth events with how many colors to display.
	EventOutgrowthManyColorsThreshold uint8
}

var DrawSelector = SpaceDrawingFilter{true,true,Distance(0),uint8(0xFF),0,}

func (filter *SpaceDrawingFilter) DisplaySettings() {
	fmt.Println("========= Space Settings =========")
	fmt.Println("Empty Nodes [N]", filter.DisplayEmptyNodes, ", Empty Connections [C]", filter.DisplayEmptyConnections)
	fmt.Println("Event Outgrowth Threshold [UP,DOWN]", filter.EventOutgrowthThreshold, ", Event Outgrowth Many Colors Threshold [U,I]", filter.EventOutgrowthManyColorsThreshold)
	fmt.Println("Event Colors Mask [1,2,3,4]", filter.EventColorMask)
}

func (filter *SpaceDrawingFilter) EventOutgrowthThresholdIncrease() {
	filter.EventOutgrowthThreshold++
	SpaceObj.createDrawingElements()
}

func (filter *SpaceDrawingFilter) EventOutgrowthThresholdDecrease() {
	filter.EventOutgrowthThreshold--
	if filter.EventOutgrowthThreshold < 0 {
		filter.EventOutgrowthThreshold = 0
	}
	SpaceObj.createDrawingElements()
}

func (filter *SpaceDrawingFilter) EventOutgrowthColorsIncrease() {
	filter.EventOutgrowthManyColorsThreshold++
	if filter.EventOutgrowthManyColorsThreshold > 4 {
		filter.EventOutgrowthManyColorsThreshold = 4
	}
	SpaceObj.createDrawingElements()
}

func (filter *SpaceDrawingFilter) EventOutgrowthColorsDecrease() {
	filter.EventOutgrowthManyColorsThreshold--
	if filter.EventOutgrowthManyColorsThreshold < 0 {
		filter.EventOutgrowthManyColorsThreshold = 0
	}
	SpaceObj.createDrawingElements()
}

func (filter *SpaceDrawingFilter) ColorMaskSwitch(color EventColor) {
	filter.EventColorMask ^= 1<<color
	SpaceObj.createDrawingElements()
}

type SpaceDrawingColor struct {
	// Bitwise flag of colors. Bits 0->red, 1->green, 2->blue, 3->yellow. If 0 then it means grey
	objColors uint8
	// The dim yes/no flag ratio to apply to each color set in above bit mask
	dimColors uint8
}

type NodeDrawingElement struct {
	objectType ObjectType
	sdc        SpaceDrawingColor
	node       *Node
}

type ConnectionDrawingElement struct {
	objectType ObjectType
	sdc        SpaceDrawingColor
	p1, p2     *Point
}

type AxeDrawingElement struct {
	objectType ObjectType
	max        int64
	neg        bool
}

func (ot ObjectType) IsAxe() bool {
	return int8(ot) >= 0 && int8(ot) <= int8(AxeZ)
}

func (ot ObjectType) IsNode() bool {
	return int8(ot) >= int8(NodeEmpty) && int8(ot) <= int8(NodeActive)
}

func (ot ObjectType) IsConnection() bool {
	return int8(ot) >= int8(Connection00)
}

func (sdc *SpaceDrawingColor) hasColor(c EventColor) bool {
	return sdc.objColors&(1<<uint8(c)) != uint8(0)
}

func (sdc *SpaceDrawingColor) hasDimm(c EventColor) bool {
	return sdc.dimColors&(1<<uint8(c)) != uint8(0)
}

func (sdc *SpaceDrawingColor) howManyColors() uint8 {
	if sdc.objColors == 0 {
		return 0
	}
	r := uint8(0)
	for c := RedEvent; c <= YellowEvent; c++ {
		if sdc.hasColor(c) {
			r++
		}
	}
	return r
}

func (sdc *SpaceDrawingColor) singleColor() EventColor {
	if sdc.objColors == 0 {
		return 0
	}
	for c := RedEvent; c <= YellowEvent; c++ {
		if sdc.hasColor(c) {
			return c
		}
	}
	return 0
}

func (sdc *SpaceDrawingColor) secondColor() EventColor {
	if sdc.objColors == 0 {
		return 0
	}
	foundOne := false
	for c := RedEvent; c <= YellowEvent; c++ {
		if sdc.hasColor(c) {
			if foundOne {
				return c
			}
			foundOne = true
		}
	}
	return 0
}

func (sdc *SpaceDrawingColor) color(blinkValue float64) int32 {
	colorSwicth := EventColor(int8(blinkValue))
	switch sdc.howManyColors() {
	case 0:
		return 0
	case 1:
		return int32(sdc.singleColor()) + 1
	case 2:
		if int(colorSwicth)%2 == 0 {
			return int32(sdc.singleColor()) + 1
		} else {
			return int32(sdc.secondColor()) + 1
		}
	case 3:
		if sdc.hasColor(colorSwicth) {
			return int32(colorSwicth) + 1
		} else {
			return 0
		}
	case 4:
		if sdc.hasColor(colorSwicth) {
			return int32(colorSwicth) + 1
		} else {
			return 0
		}
	}
	return 0
}

func (sdc *SpaceDrawingColor) dimmer(blinkValue float64) float32 {
	colorSwicth := EventColor(int8(blinkValue))
	switch sdc.howManyColors() {
	case 0:
		return defaultGreyDimmer
	case 1:
		if sdc.hasDimm(sdc.singleColor()) {
			return defaultOldEventDimmer
		}
		return noDimmer
	case 2:
		if int(colorSwicth)%2 == 0 {
			if sdc.hasDimm(sdc.singleColor()) {
				return defaultOldEventDimmer
			}
			return noDimmer
		} else {
			if sdc.hasDimm(sdc.secondColor()) {
				return defaultOldEventDimmer
			}
			return noDimmer
		}
	case 3:
		if sdc.hasColor(colorSwicth) {
			if sdc.hasDimm(colorSwicth) {
				return defaultOldEventDimmer
			}
			return noDimmer
		} else {
			return defaultGreyDimmer
		}
	case 4:
		if sdc.hasColor(colorSwicth) {
			if sdc.hasDimm(colorSwicth) {
				return defaultOldEventDimmer
			}
			return noDimmer
		} else {
			return defaultGreyDimmer
		}
	}
	return defaultGreyDimmer
}

func MakeNodeDrawingElement(node *Node) *NodeDrawingElement {
	// Collect all the colors of event outgrowth of this node. Dim if not latest
	isActive := false
	sdc := SpaceDrawingColor{}
	for c := RedEvent; c <= YellowEvent; c++ {
		for _, eo := range node.outgrowths {
			if eo.IsActive(DrawSelector.EventOutgrowthThreshold) {
				sdc.objColors |= 1 << uint8(eo.event.color)
				// TODO: Another threshold for dim?
				//if eo.state != EventOutgrowthLatest && !eo.IsRoot() {
				//	sdc.dimColors |= 1 << uint8(eo.event.color)
				//}
				isActive = true
			}
		}
	}
	if isActive {
		return &NodeDrawingElement{
			NodeActive, sdc, node,
		}
	} else {
		return &NodeDrawingElement{
			NodeEmpty, sdc, node,
		}
	}

}

func MakeConnectionDrawingElement(conn *Connection) *ConnectionDrawingElement {
	n1 := conn.N1
	n2 := conn.N2
	// Collect all the colors of latest event outgrowth of a node coming from the other node
	sdc := SpaceDrawingColor{}
	for c := RedEvent; c <= YellowEvent; c++ {
		for _, eo1 := range n1.outgrowths {
			if eo1.IsActive(DrawSelector.EventOutgrowthThreshold) {
				if eo1.from != nil && eo1.from.node == n2 && eo1.from.node.IsActive(DrawSelector.EventOutgrowthThreshold) {
					sdc.objColors |= 1 << uint8(eo1.event.color)
				}
			}
		}
		for _, eo2 := range n2.outgrowths {
			if eo2.IsActive(DrawSelector.EventOutgrowthThreshold) {
				if eo2.from != nil && eo2.from.node == n1 && eo2.from.node.IsActive(DrawSelector.EventOutgrowthThreshold) {
					sdc.objColors |= 1 << uint8(eo2.event.color)
				}
			}
		}
	}
	p1 := n1.point
	p2 := n2.point
	cd := GetConnectionDetails(*p1, *p2)
	if cd.ConnNeg {
		return &ConnectionDrawingElement{ObjectType(uint8(Connection00) + cd.ConnNumber), sdc, p2, p1,}
	}
	return &ConnectionDrawingElement{ObjectType(uint8(Connection00) + cd.ConnNumber), sdc, p1, p2,}
}

// NodeDrawingElement functions
func (n *NodeDrawingElement) Key() ObjectType {
	return n.objectType
}

func (n *NodeDrawingElement) Display() bool {
	if n.objectType == NodeActive {
		if n.node.IsRoot() {return true}
		return n.sdc.objColors&DrawSelector.EventColorMask != 0 && n.sdc.howManyColors() >= DrawSelector.EventOutgrowthManyColorsThreshold
	}
	return DrawSelector.DisplayEmptyNodes
}

func (n *NodeDrawingElement) Color(blinkValue float64) int32 {
	return n.sdc.color(blinkValue)
}

func (n *NodeDrawingElement) Dimmer(blinkValue float64) float32 {
	return n.sdc.dimmer(blinkValue)
}

func (n *NodeDrawingElement) Pos() *Point {
	return n.node.point
}

// ConnectionDrawingElement functions
func (c *ConnectionDrawingElement) Key() ObjectType {
	return c.objectType
}

func (c *ConnectionDrawingElement) Display() bool {
	if c.sdc.objColors&DrawSelector.EventColorMask != uint8(0) && c.sdc.howManyColors() >= DrawSelector.EventOutgrowthManyColorsThreshold {
		return true
	}
	return DrawSelector.DisplayEmptyConnections
}

func (c *ConnectionDrawingElement) Color(blinkValue float64) int32 {
	return c.sdc.color(blinkValue)
}

func (c *ConnectionDrawingElement) Dimmer(blinkValue float64) float32 {
	dimmer := c.sdc.dimmer(blinkValue)
	return dimmer
}

func (c *ConnectionDrawingElement) Pos() *Point {
	return c.p1
}

// AxeDrawingElement functions
func (a *AxeDrawingElement) Key() ObjectType {
	return a.objectType
}

func (a *AxeDrawingElement) Display() bool {
	return true
}

func (a *AxeDrawingElement) Color(blinkValue float64) int32 {
	switch a.objectType {
	case AxeX:
		return int32(RedEvent) + 1
	case AxeY:
		return int32(GreenEvent) + 1
	case AxeZ:
		return int32(BlueEvent) + 1
	}
	return 0
}

func (a *AxeDrawingElement) Dimmer(blinkValue float64) float32 {
	return 1.0
}

func (a *AxeDrawingElement) Pos() *Point {
	if a.neg {
		switch a.objectType {
		case AxeX:
			return &Point{-a.max, 0, 0}
		case AxeY:
			return &Point{0, -a.max, 0}
		case AxeZ:
			return &Point{0, 0, -a.max}
		}
	}
	return &Origin
}
