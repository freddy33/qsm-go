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

func (ot ObjectType) IsAxe() bool {
	return int16(ot) >= 0 && int16(ot) <= int16(AxeZ)
}

func (ot ObjectType) IsNode() bool {
	return int16(ot) >= int16(Node0) && int16(ot) <= int16(NodeC)
}

func (ot ObjectType) IsConnection() bool {
	return int16(ot) >= int16(Connection1) && int16(ot) <= int16(Connection6)
}

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

func DS(p1, p2 *Point) int64 {
	x := p2.X() - p1.X()
	y := p2.Y() - p1.Y()
	z := p2.Z() - p1.Z()
	return x*x + y*y + z*z
}

func (p *Point) GetNodeId() NodeID {
	return NodeID(p.X()+1000 + (p.Y()+1000)*1000 + (p.Z()+1000)*1000000)
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

func (p *Point) IsBorder(max int64) bool {
	for _, c := range p {
		if c > 0 && c >= max-1 {
			return true
		}
		if c < 0 && c <= -max+1 {
			return true
		}
	}
	return false
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
	for x := -maxByThree; x <= maxByThree; x++ {
		for y := -maxByThree; y <= maxByThree; y++ {
			for z := -maxByThree; z <= maxByThree; z++ {
				p := Point{x * THREE, y * THREE, z * THREE}
				s.GetNode(p)
			}
		}
	}
	// All nodes that are not main with nil connections find good one
	for _, node := range s.nodes {
		if !node.P.IsMainPoint() && node.HasFreeConnections() {
			for i := 1; i < 3; i++ {
				if node.C[i] == nil {
					for _, other := range s.nodes {
						if node != other && !other.P.IsMainPoint() && other.HasFreeConnections() && DS(other.P, node.P) == 3 {
							node.C[i] = other
							for j := 0; j < 3; j++ {
								if other.C[j] == nil {
									other.C[j] = node
									break
								}
							}
							break
						}
					}
				}
			}
		}
	}

	// Verify all connections length squared is 2 or 3
	for _, node := range s.nodes {
		for i, c := range node.C {
			if c != nil {
				ds := DS(node.P, c.P)
				if ds != 2 && ds != 3 {
					fmt.Println("something wrong with node",node.P,ds,"connection",i,c.P)
				}
			} else {
				// Should be on the border
				if !node.P.IsBorder(max) {
					fmt.Println("something wrong with node connection not done for",node.P,"connection",i)
				}
			}
		}
	}
}

func (n *Node) HasFreeConnections() bool {
	for _, c := range n.C {
		if c == nil {
			return true
		}
	}
	return false
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
	bv := p2.Sub(*p1)
	if p1.IsMainPoint() {
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1)+i), p1, p2,}
			}
		}
		fmt.Println("What 1",p1,p2,bv)
		return &ConnectionDrawingElement{Connection1, p1, p2,}
	} else if p2.IsMainPoint() {
		bv = bv.Mul(-1)
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1)+i), p2, p1,}
			}
		}
		fmt.Println("What 2",p1,p2,bv)
		return &ConnectionDrawingElement{Connection1, p2, p1,}
	} else {
		if bv[0] == 1 {
			if bv[1] != -1 || bv[2] != -1 {
				fmt.Println("What 3",p1,p2,bv)
			}
			return &ConnectionDrawingElement{Connection4, p1, p2,}
		} else {
			if bv[0] != -1 || bv[1] != 1 || bv[2] != 1 {
				fmt.Println("What 4",p1,p2,bv)
			}
			return &ConnectionDrawingElement{Connection5, p1, p2,}
		}
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
