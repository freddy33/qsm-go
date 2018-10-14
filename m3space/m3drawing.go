package m3space

import "fmt"

type ObjectType int8

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
	Key() ObjectType
	Color() ObjectColor
	Alpha() float32
	Pos() *Point
}

type NodeDrawingElement struct {
	t ObjectType
	c ObjectColor
	a float32
	n *Node
}

type ConnectionDrawingElement struct {
	t      ObjectType
	c      ObjectColor
	a      float32
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

func MakeConnectionDrawingElement(p1, p2 *Point) *ConnectionDrawingElement {
	bv := p2.Sub(*p1)
	if p1.IsMainPoint() {
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1) + i), Grey, 0.7, p1, p2,}
			}
		}
		fmt.Println("What 1", p1, p2, bv)
		return &ConnectionDrawingElement{Connection1, Grey, 0.7, p1, p2,}
	} else if p2.IsMainPoint() {
		fmt.Println("What 2", p1, p2, bv)
		return &ConnectionDrawingElement{Connection1, Grey, 0.7, p2, p1,}
	} else {
		if bv[0] == 1 {
			if bv[1] != -1 || bv[2] != -1 {
				fmt.Println("What 3", p1, p2, bv)
			}
			return &ConnectionDrawingElement{Connection4, Grey, 0.7, p1, p2,}
		} else {
			if bv[0] != -1 || bv[1] != 1 || bv[2] != 1 {
				fmt.Println("What 4", p1, p2, bv)
			}
			return &ConnectionDrawingElement{Connection5, Grey, 0.7, p1, p2,}
		}
	}
}

// NodeDrawingElement functions
func (n *NodeDrawingElement) Key() ObjectType {
	return n.t
}

func (n *NodeDrawingElement) Color() ObjectColor {
	return n.c
}

func (n *NodeDrawingElement) Alpha() float32 {
	return n.a
}

func (n *NodeDrawingElement) Pos() *Point {
	return n.n.P
}

// ConnectionDrawingElement functions
func (c *ConnectionDrawingElement) Key() ObjectType {
	return c.t
}

func (c *ConnectionDrawingElement) Color() ObjectColor {
	return c.c
}

func (c *ConnectionDrawingElement) Alpha() float32 {
	return c.a
}

func (c *ConnectionDrawingElement) Pos() *Point {
	return c.p1
}

// AxeDrawingElement functions
func (a *AxeDrawingElement) Key() ObjectType {
	return a.t
}

func (a *AxeDrawingElement) Alpha() float32 {
	return 1.0
}

func (a *AxeDrawingElement) Color() ObjectColor {
	switch a.t {
	case AxeX:
		return Red
	case AxeY:
		return Green
	case AxeZ:
		return Blue
	}
	return Grey
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
