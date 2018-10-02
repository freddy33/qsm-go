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

type NodeID int64

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
	nodes    map[NodeID]*Node
	events   map[EventID]*Event
	current  TickTime
	max      int64
	Elements []SpaceDrawingElement
}

var SpaceObj = Space{
	make(map[NodeID]*Node),
	make(map[EventID]*Event),
	0,
	9,
	[]SpaceDrawingElement{},
}

type ObjectType int16

const (
	AxeX        ObjectType = iota
	AxeY
	AxeZ
	Node0
	NodeA
	NodeB
	NodeC
	Connection1
	Connection2
	Connection3
	Connection4
	Connection5
	Connection6
)

const THREE = 3

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}
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

func (p *Point) GetNodeId() NodeID {
	return NodeID(p.X() + p.Y()*10000 + p.Z()*100000000)
}

func (p *Point) IsMainPoint() bool {
	allDivByThree := true
	for _, c := range p {
		if c%THREE != 0 {
			allDivByThree = false
		}
	}
	return allDivByThree
}

func (s *Space) GetNode(p Point) *Node {
	nId := p.GetNodeId()
	n, ok := s.nodes[nId]
	if ok {
		return n
	}
	newNode := Node{}
	newNode.P = &p
	s.nodes[nId] = &newNode
	if p.IsMainPoint() {
		s.connectToBasePoints(nId)
	}
	return &newNode
}

func (s *Space) connectToBasePoints(nId NodeID) {
	bn, ok := s.nodes[nId]
	if !ok {
		fmt.Println("Passing Node Id", nId, "does exists in map")
		return
	}
	if !bn.P.IsMainPoint() {
		fmt.Println("Passing Node Id", nId, "is not a main point")
		return
	}
	for i, b := range BasePoints {
		bpn := s.GetNode(bn.P.Add(b))
		bn.C[i] = bpn
		bpn.C[0] = bn
	}
}

func (s *Space) CreateNodes(max int64) {
	s.GetNode(Origin)
	maxByThree := int64(max / THREE)
	fmt.Println("Max by three", maxByThree)
	for x := -maxByThree; x < maxByThree; x++ {
		for y := -maxByThree; y < maxByThree; y++ {
			for z := -maxByThree; z < maxByThree; z++ {
				p := Point{x * THREE, y * THREE, z * THREE}
				s.GetNode(p)
			}
		}
	}
}

func (s *Space) CreateStuff(max int64) {
	s.CreateNodes(max)
	e := Event{0, s.GetNode(Origin), 0, EventA,}
	s.events[0] = &e
	s.createDrawingElements()
}

func (s *Space) createDrawingElements() {
	elements := make([]SpaceDrawingElement, 6+len(s.nodes)+3*len(s.nodes)+len(s.events))
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
	for _, node := range s.nodes {
		elements[offset] = &NodeDrawingElement{
			Node0,
			node,
		}
		offset++
		for c := 0; c < THREE; c++ {
			if node.C[c] != nil {
				elements[offset] = MakeConnectionDrawingElement(node.P, node.C[c].P)
			} else {
				elements[offset] = nil
			}
			offset++
		}
	}
	for _, evt := range s.events {
		elements[offset] = &NodeDrawingElement{
			ObjectType(int16(NodeA) + int16(evt.K)),
			evt.N,
		}
	}
	fmt.Println("Created", len(elements), "elements.")
	s.Elements = elements
}

type SpaceDrawingElement interface {
	Key() ObjectType
	Pos() *Point
}

type NodeDrawingElement struct {
	t ObjectType
	n *Node
}

type ConnectionDrawingElement struct {
	t      ObjectType
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
	return n.n.P
}

// ConnectionDrawingElement functions
func MakeConnectionDrawingElement(p1, p2 *Point) *ConnectionDrawingElement {
	if p1.IsMainPoint() {
		bv := p2.Sub(*p1)
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1)+i), p1, p2,}
			}
		}
		return &ConnectionDrawingElement{Connection1, p1, p2,}
	} else if p2.IsMainPoint() {
		bv := p1.Sub(*p2)
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1)+i), p2, p1,}
			}
		}
		return &ConnectionDrawingElement{Connection1, p2, p1,}
	} else {
		return &ConnectionDrawingElement{Connection4, p1, p2,}
	}
}

func (c *ConnectionDrawingElement) Key() ObjectType {
	return c.t
}

func (c *ConnectionDrawingElement) Pos() *Point {
	return c.p1
}

// AxeDrawingElement functions
func (a *AxeDrawingElement) Key() ObjectType {
	return a.t
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
