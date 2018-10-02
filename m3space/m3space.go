package m3space

import "fmt"

type Point [3]int64

func (p Point) X() int64 {
	return p[0]
}

func (p Point) Y() int64 {
	return p[1]
}

func (p Point) Z() int64 {
	return p[2]
}

type Node struct {
	P *Point
	C [3]*Node
}

type TickTime uint64

type Distance uint64

type EventID uint64

type EventKind int8

const (
	EventA EventKind = iota
	EventB
	EventC
)

type Event struct {
	ID EventID
	N  *Node
	T  TickTime
	K  EventKind
}

type EventOutgrowth struct {
	Evt   *Event
	D     Distance
	Nodes []*Node
}

type Space struct {
	events   map[EventID]*Event
	current  TickTime
	max      int64
	Elements []SpaceDrawingElement
}

var SpaceObj = Space{
	make(map[EventID]*Event),
	0,
	9,
	[]SpaceDrawingElement{},
}

type ObjectType int16

const (
	AxeX       ObjectType = iota
	AxeY
	AxeZ
	Node0
	NodeA
	NodeB
	NodeC
	Connection1
	Connection2
)

const THREE = 3

var Origin = Point{0, 0, 0}
var BasePoints = [3]Point{{1, 1, 0}, {0, -1, 1}, {-1, 0, -1}}

func (p Point) Mul(m int64) Point {
	return Point{p[0] * m, p[1] * m, p[2] * m}
}

func (p1 Point) Add(p2 Point) Point {
	return Point{p1[0] + p2[0], p1[1] + p2[1], p1[2] + p2[2]}
}

func (p1 Point) Sub(p2 Point) Point {
	return Point{p1[0] - p2[0], p1[1] - p2[1], p1[2] - p2[2]}
}

func DS(p1, p2 Point) int64 {
	return (p2[0] - p1[0]) ^ 2 + (p2[1] - p1[1]) ^ 2 + (p2[2] - p1[2]) ^ 2
}

func (s *Space) CreateStuff(max int64) {
	n := Node{&Origin, [3]*Node{}, }
	e := Event{0, &n, 0, EventA, }
	s.events[0] = &e
	for i, b := range BasePoints {
		nn := Node{&b, [3]*Node{}}
		nn.C[0] = &n
		n.C[i] = &nn
		ee := Event{EventID(1+i), &nn, 0, EventA, }
		s.events[EventID(1+i)] = &ee
	}
	s.createDrawingElements()
}

func (s *Space) createDrawingElements() {
	elements := make([]SpaceDrawingElement, 6+len(s.events)+3*len(s.events))
	offset := 0
	for axe := 0; axe < 3; axe++ {
		elements[offset] = &AxeDrawingElement{
			ObjectType(axe),
			s.max,
			false,
		}
		offset++
		elements[offset] = &AxeDrawingElement{
			ObjectType(axe),
			s.max,
			true,
		}
		offset++
	}
	for _, evt := range s.events {
		elements[offset] = &NodeDrawingElement{
			ObjectType(int16(NodeA) + int16(evt.K)),
			evt,
		}
		offset++
		for c := 0; c < THREE; c++ {
			if evt.N.C[c] != nil {
				elements[offset] = MakeConnectionDrawingElement(evt.N.P, evt.N.C[c].P)
			} else {
				elements[offset] = nil
			}
			offset++
		}
	}
	fmt.Println("Created",len(elements),"elements. Keys are:")
	for _, el := range elements {
		if el != nil {
			fmt.Println(el.Key())
		} else {
			fmt.Println("nil")
		}
	}
	s.Elements = elements
}

type SpaceDrawingElement interface {
	Key() ObjectType
	Pos() *Point
	EndPoint() *Point
}

type NodeDrawingElement struct {
	t   ObjectType
	evt *Event
}

type ConnectionDrawingElement struct {
	t   ObjectType
	p1, p2 *Point
}

type AxeDrawingElement struct {
	t   ObjectType
	max int64
	neg bool
}

// NodeDrawingElement functions
func (n *NodeDrawingElement) Key() ObjectType {
	return n.t
}

func (n *NodeDrawingElement) Pos() *Point {
	return n.evt.N.P
}

func (n *NodeDrawingElement) EndPoint() *Point {
	return nil
}

// ConnectionDrawingElement functions
func MakeConnectionDrawingElement(p1, p2 *Point) *ConnectionDrawingElement {
	divByThree := false
	for _, c := range p1 {
		if c % THREE == 0 {
			divByThree = true
		}
	}
	for _, c := range p2 {
		if c % THREE == 0 {
			divByThree = true
		}
	}
	if divByThree {
		return &ConnectionDrawingElement{Connection1, p1, p2,}
	} else {
		return &ConnectionDrawingElement{Connection2, p1, p2,}
	}
}

func (c *ConnectionDrawingElement) Key() ObjectType {
	return c.t
}

func (c *ConnectionDrawingElement) Pos() *Point {
	return c.p1
}

func (c *ConnectionDrawingElement) EndPoint() *Point {
	return c.p2
}

// AxeDrawingElement functions
func (a *AxeDrawingElement) Key() ObjectType {
	return a.t
}

func (a *AxeDrawingElement) Pos() *Point {
	return &Origin
}

func (a *AxeDrawingElement) EndPoint() *Point {
	val := a.max
	if a.neg {
		val = -val
	}
	switch a.t {
	case AxeX:
		return &Point{val, 0, 0}
	case AxeY:
		return &Point{0, val, 0}
	case AxeZ:
		return &Point{0, 0, val}
	}
	fmt.Println("Type is not an Axe type but", a.t)
	return nil
}
