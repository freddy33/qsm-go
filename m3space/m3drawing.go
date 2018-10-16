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
	Connection1
	Connection2
	Connection3
	Connection4
	Connection5
	Connection6
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

type SpaceDrawingColor struct {
	// Bitwise flag of colors. Bits 0->red, 1->green, 2->blue, 3->yellow. If 0 then it means grey
	objColors uint8
	// The dim yes/no flag ratio to apply to each color set in above bit mask
	dimColors uint8
}

type NodeDrawingElement struct {
	t ObjectType
	c SpaceDrawingColor
	n *Node
}

type ConnectionDrawingElement struct {
	t      ObjectType
	c      SpaceDrawingColor
	p1, p2 *Point
}

type AxeDrawingElement struct {
	t   ObjectType
	max int64
	neg bool
}

func (ot ObjectType) IsAxe() bool {
	return int8(ot) >= 0 && int8(ot) <= int8(AxeZ)
}

func (ot ObjectType) IsNode() bool {
	return int8(ot) >= int8(NodeEmpty) && int8(ot) <= int8(NodeActive)
}

func (ot ObjectType) IsConnection() bool {
	return int8(ot) >= int8(Connection1) && int8(ot) <= int8(Connection6)
}

func (sdc *SpaceDrawingColor) hasColor(c EventColor) bool {
	return sdc.objColors&(1<<uint8(c)) != uint8(0)
}

func (sdc *SpaceDrawingColor) hasDimm(c EventColor) bool {
	return sdc.dimColors&(1<<uint8(c)) != uint8(0)
}

func (sdc *SpaceDrawingColor) howManyColors() int8 {
	if sdc.objColors == 0 {
		return 0
	}
	r := int8(0)
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
		return int32(sdc.singleColor())+1
	case 2:
		if int(colorSwicth)%2 == 0 {
			return int32(sdc.singleColor())+1
		} else {
			return int32(sdc.secondColor())+1
		}
	case 3:
		if sdc.hasColor(colorSwicth) {
			return int32(colorSwicth)+1
		} else {
			return 0
		}
	case 4:
		if sdc.hasColor(colorSwicth) {
			return int32(colorSwicth)+1
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
		for _, eo := range node.E {
			if eo != nil && (eo.DistanceFromLatest() <= 1 || eo.IsRoot()) {
				sdc.objColors |= 1 << uint8(eo.event.color)
				// Event root themselves never dim
				if eo.state != EventOutgrowthLatest && eo.IsRoot() {
					sdc.dimColors |= 1 << uint8(eo.event.color)
				}
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
		for _, eo1 := range n1.E {
			if eo1.state == EventOutgrowthLatest {
				if eo1.from != nil && eo1.from.node == n2 {
					sdc.objColors |= 1 << uint8(eo1.event.color)
				}
			}
		}
		for _, eo2 := range n2.E {
			if eo2.state == EventOutgrowthLatest {
				if eo2.from != nil && eo2.from.node == n1 {
					sdc.objColors |= 1 << uint8(eo2.event.color)
				}
			}
		}
	}
	p1 := n1.P
	p2 := n2.P
	bv := p2.Sub(*p1)
	if p1.IsMainPoint() {
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1) + i), sdc, p1, p2,}
			}
		}
		fmt.Println("Not possible! Connection from", p1, p2, "has p1 main and make vector", bv, "which is not part of any base point vector!")
		return nil
	} else if p2.IsMainPoint() {
		fmt.Println("Not possible! Connection from", p1, p2, "has P2 has main point which is not possible!")
		return nil
	} else {
		if bv[0] == 1 {
			if bv[1] != -1 || bv[2] != -1 {
				fmt.Println("Not possible! Connection from", p1, p2, "make vector", bv, "which is a DS=3 connection with X value 1 andr Y or Z value not neg 1!")
				return nil
			}
			return &ConnectionDrawingElement{Connection4, sdc, p1, p2,}
		} else {
			if bv[0] != -1 || bv[1] != 1 || bv[2] != 1 {
				fmt.Println("Not possible! Connection from", p1, p2, "make vector", bv, "which is a DS=3 connection with X not value 1 so should be (-1,1,1)!")
				return nil
			}
			return &ConnectionDrawingElement{Connection5, sdc, p1, p2,}
		}
	}
}

// NodeDrawingElement functions
func (n *NodeDrawingElement) Key() ObjectType {
	return n.t
}

func (n *NodeDrawingElement) Display() bool {
	if n.t == NodeActive {
		return true
	}
	return false
}

func (n *NodeDrawingElement) Color(blinkValue float64) int32 {
	return n.c.color(blinkValue)
}

func (n *NodeDrawingElement) Dimmer(blinkValue float64) float32 {
	return n.c.dimmer(blinkValue)
}

func (n *NodeDrawingElement) Pos() *Point {
	return n.n.P
}

// ConnectionDrawingElement functions
func (c *ConnectionDrawingElement) Key() ObjectType {
	return c.t
}

func (c *ConnectionDrawingElement) Display() bool {
	if c.c.objColors != uint8(0) {
		return true
	}
	return false
}

func (c *ConnectionDrawingElement) Color(blinkValue float64) int32 {
	return c.c.color(blinkValue)
}

func (c *ConnectionDrawingElement) Dimmer(blinkValue float64) float32 {
	dimmer := c.c.dimmer(blinkValue)
	if dimmer < 1.0 {
		dimmer *= 0.5
	}
	return dimmer
}

func (c *ConnectionDrawingElement) Pos() *Point {
	return c.p1
}

// AxeDrawingElement functions
func (a *AxeDrawingElement) Key() ObjectType {
	return a.t
}

func (a *AxeDrawingElement) Display() bool {
	return true
}

func (a *AxeDrawingElement) Color(blinkValue float64) int32 {
	switch a.t {
	case AxeX:
		return int32(RedEvent)+1
	case AxeY:
		return int32(GreenEvent)+1
	case AxeZ:
		return int32(BlueEvent)+1
	}
	return 0
}

func (a *AxeDrawingElement) Dimmer(blinkValue float64) float32 {
	return 1.0
}

func (a *AxeDrawingElement) Pos() *Point {
	if a.neg {
		switch a.t {
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
