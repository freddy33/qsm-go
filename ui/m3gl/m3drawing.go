package m3gl

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

const (
	noDimmer              = 1.0
	defaultGreyDimmer     = 0.7
	defaultOldEventDimmer = 0.3
)

type ObjectType uint8

const (
	AxeX ObjectType = iota
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
	Pos() *m3point.Point
	// Return the obj_color int for the shader program
	Color(blinkValue float64) int32
	// Return the obj_dimmer int for the shader program
	Dimmer(blinkValue float64) float32
	// Display flag
	Display(filter SpaceDrawingFilter) bool
}

type SpaceDrawingFilter struct {
	// Display grey empty nodes or not
	DisplayEmptyNodes bool
	// Display grey empty connections or not
	DisplayEmptyConnections bool
	// The events of certain colors to display. This is a mask.
	EventColorMask uint8
	// The outgrowth events with how many colors to display.
	EventOutgrowthManyColorsThreshold uint8
	// The space the filter apply to
	SpaceTime m3space.SpaceTimeIfc
}

func (filter *SpaceDrawingFilter) DisplaySettings() {
	fmt.Println("========= SpaceTime Settings =========")
	fmt.Println("Empty Nodes [N]", filter.DisplayEmptyNodes, ", Empty Connections [C]", filter.DisplayEmptyConnections)
	fmt.Println("Event Outgrowth Threshold [UP,DOWN]", filter.SpaceTime.GetSpace().GetActiveThreshold(), ", Event Outgrowth Many Colors Threshold [U,I]", filter.EventOutgrowthManyColorsThreshold)
	fmt.Println("Event Colors Mask [1,2,3,4]", filter.EventColorMask)
}

func (filter *SpaceDrawingFilter) EventOutgrowthColorsIncrease() {
	filter.EventOutgrowthManyColorsThreshold++
	if filter.EventOutgrowthManyColorsThreshold > 4 {
		filter.EventOutgrowthManyColorsThreshold = 4
	}
}

func (filter *SpaceDrawingFilter) EventOutgrowthColorsDecrease() {
	filter.EventOutgrowthManyColorsThreshold--
	if filter.EventOutgrowthManyColorsThreshold < 0 {
		filter.EventOutgrowthManyColorsThreshold = 0
	}
}

func (filter *SpaceDrawingFilter) ColorMaskSwitch(color m3space.EventColor) {
	filter.EventColorMask ^= 1 << color
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
	node       m3space.SpaceTimeNodeIfc
}

type ConnectionDrawingElement struct {
	objectType ObjectType
	sdc        SpaceDrawingColor
	pos        *m3point.Point
}

type AxeDrawingElement struct {
	objectType ObjectType
	max        m3point.CInt
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

func (sdc *SpaceDrawingColor) hasColor(c m3space.EventColor) bool {
	return sdc.objColors&uint8(c) != uint8(0)
}

func (sdc *SpaceDrawingColor) hasDimming(c m3space.EventColor) bool {
	return sdc.dimColors&(1<<uint8(c)) != uint8(0)
}

func (sdc *SpaceDrawingColor) howManyColors() uint8 {
	if sdc.objColors == 0 {
		return 0
	}
	r := uint8(0)
	for _, c := range m3space.AllColors {
		if sdc.hasColor(c) {
			r++
		}
	}
	return r
}

func (sdc *SpaceDrawingColor) singleColor() m3space.EventColor {
	if sdc.objColors == 0 {
		return 0
	}
	for _, c := range m3space.AllColors {
		if sdc.hasColor(c) {
			return c
		}
	}
	return 0
}

func (sdc *SpaceDrawingColor) secondColor() m3space.EventColor {
	if sdc.objColors == 0 {
		return 0
	}
	foundOne := false
	for _, c := range m3space.AllColors {
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
	colorSwitch := uint8(blinkValue)
	color := m3space.EventColor(1 << colorSwitch)
	switch sdc.howManyColors() {
	case 0:
		return 0
	case 1:
		return int32(sdc.singleColor())
	case 2:
		if int(colorSwitch)%2 == 0 {
			return int32(sdc.singleColor())
		} else {
			return int32(sdc.secondColor())
		}
	case 3:
		if sdc.hasColor(color) {
			return int32(color)
		} else {
			return 0
		}
	case 4:
		if sdc.hasColor(color) {
			return int32(color)
		} else {
			return 0
		}
	}
	return 0
}

func (sdc *SpaceDrawingColor) dimmer(blinkValue float64) float32 {
	colorSwitch := uint8(blinkValue)
	color := m3space.EventColor(1 << colorSwitch)
	switch sdc.howManyColors() {
	case 0:
		return defaultGreyDimmer
	case 1:
		if sdc.hasDimming(sdc.singleColor()) {
			return defaultOldEventDimmer
		}
		return noDimmer
	case 2:
		if int(colorSwitch)%2 == 0 {
			if sdc.hasDimming(sdc.singleColor()) {
				return defaultOldEventDimmer
			}
			return noDimmer
		} else {
			if sdc.hasDimming(sdc.secondColor()) {
				return defaultOldEventDimmer
			}
			return noDimmer
		}
	case 3:
		if sdc.hasColor(color) {
			if sdc.hasDimming(color) {
				return defaultOldEventDimmer
			}
			return noDimmer
		} else {
			return defaultGreyDimmer
		}
	case 4:
		if sdc.hasColor(color) {
			if sdc.hasDimming(color) {
				return defaultOldEventDimmer
			}
			return noDimmer
		} else {
			return defaultGreyDimmer
		}
	}
	return defaultGreyDimmer
}

func MakeNodeDrawingElement(node m3space.SpaceTimeNodeIfc) *NodeDrawingElement {
	// Collect all the colors of event outgrowth of this node. Dim if not latest
	sdc := SpaceDrawingColor{}
	sdc.objColors = node.GetColorMask()
	// TODO: Another threshold for dim?
	//if eo.state != EventOutgrowthLatest && !eo.IsRoot() {
	//	sdc.dimColors |= 1 << uint8(eo.event.color)
	//}

	if sdc.objColors != uint8(0) {
		return &NodeDrawingElement{
			NodeActive, sdc, node,
		}
	} else {
		return &NodeDrawingElement{
			NodeEmpty, sdc, node,
		}
	}

}

func MakeConnectionDrawingElement(node m3space.SpaceTimeNodeIfc, point m3point.Point, connId m3point.ConnectionId) *ConnectionDrawingElement {
	// Collect all the colors of latest event outgrowth of a node coming from the other node
	sdc := SpaceDrawingColor{}
	// Take the color of the source. TODO: Not true should be & on src and target
	sdc.objColors = node.GetColorMask()
	return &ConnectionDrawingElement{getConnectionObjectType(connId), sdc, &point}
}

func getConnectionObjectType(cdId m3point.ConnectionId) ObjectType {
	if cdId > 0 {
		return ObjectType(m3point.ConnectionId(Connection00) + cdId*2)
	} else {
		return ObjectType(m3point.ConnectionId(Connection00) + 1 - cdId*2)
	}
}

// NodeDrawingElement functions
func (n NodeDrawingElement) Key() ObjectType {
	return n.objectType
}

func (n NodeDrawingElement) Display(filter SpaceDrawingFilter) bool {
	if n.objectType == NodeActive {
		if n.node.HasRoot() {
			return true
		}
		return n.sdc.objColors&filter.EventColorMask != 0 && n.sdc.howManyColors() >= filter.EventOutgrowthManyColorsThreshold
	}
	return filter.DisplayEmptyNodes
}

func (n NodeDrawingElement) Color(blinkValue float64) int32 {
	return n.sdc.color(blinkValue)
}

func (n NodeDrawingElement) Dimmer(blinkValue float64) float32 {
	return n.sdc.dimmer(blinkValue)
}

func (n NodeDrawingElement) Pos() *m3point.Point {
	p, err := n.node.GetPoint()
	if err != nil {
		Log.Error(err)
		return nil
	}
	return p
}

// ConnectionDrawingElement functions
func (c ConnectionDrawingElement) Key() ObjectType {
	return c.objectType
}

func (c ConnectionDrawingElement) Display(filter SpaceDrawingFilter) bool {
	if c.sdc.objColors&filter.EventColorMask != uint8(0) && c.sdc.howManyColors() >= filter.EventOutgrowthManyColorsThreshold {
		return true
	}
	return filter.DisplayEmptyConnections
}

func (c ConnectionDrawingElement) Color(blinkValue float64) int32 {
	return c.sdc.color(blinkValue)
}

func (c ConnectionDrawingElement) Dimmer(blinkValue float64) float32 {
	dimmer := c.sdc.dimmer(blinkValue)
	return dimmer
}

func (c ConnectionDrawingElement) Pos() *m3point.Point {
	return c.pos
}

// AxeDrawingElement functions
func (a AxeDrawingElement) Key() ObjectType {
	return a.objectType
}

func (a AxeDrawingElement) Display(filter SpaceDrawingFilter) bool {
	return true
}

func (a AxeDrawingElement) Color(blinkValue float64) int32 {
	switch a.objectType {
	case AxeX:
		return int32(m3space.RedEvent)
	case AxeY:
		return int32(m3space.GreenEvent)
	case AxeZ:
		return int32(m3space.BlueEvent)
	}
	return 0
}

func (a AxeDrawingElement) Dimmer(blinkValue float64) float32 {
	return 1.0
}

func (a AxeDrawingElement) Pos() *m3point.Point {
	if a.neg {
		switch a.objectType {
		case AxeX:
			return &m3point.Point{-a.max, 0, 0}
		case AxeY:
			return &m3point.Point{0, -a.max, 0}
		case AxeZ:
			return &m3point.Point{0, 0, -a.max}
		}
	}
	return &m3point.Origin
}
