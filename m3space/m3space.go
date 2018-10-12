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
	E []*EventOutgrowth
	C [3]*Connection
}

type Connection struct {
	N1, N2 *Node
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
	O  []*EventOutgrowth
}

type EventOutgrowth struct {
	N *Node
	E *Event
	D Distance
}

type Space struct {
	nodes       []*Node
	connections []*Connection
	events      map[EventID]*Event
	current     TickTime
	max         int64
	Elements    []SpaceDrawingElement
}

var SpaceObj = Space{
	make([]*Node, 0, 108),
	make([]*Connection, 0, 500),
	make(map[EventID]*Event),
	0,
	9,
	make([]SpaceDrawingElement, 0, 500),
}

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

type ObjectColor int8

const (
	Grey  ObjectColor = iota
	Red
	Green
	Blue
	Yellow
)

func (ot ObjectType) IsAxe() bool {
	return int8(ot) >= 0 && int8(ot) <= int8(AxeZ)
}

func (ot ObjectType) IsNode() bool {
	return int8(ot) >= int8(NodeEmpty) && int8(ot) <= int8(NodeActive)
}

func (ot ObjectType) IsConnection() bool {
	return int8(ot) >= int8(Connection1) && int8(ot) <= int8(Connection6)
}

const THREE = 3

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}
var BasePoints = [3]Point{{1, 1, 0}, {0, -1, 1}, {-1, 0, -1}}

func (p *Point) Mul(m int64) Point {
	return Point{p[0] * m, p[1] * m, p[2] * m}
}

func (p1 *Point) Add(p2 Point) Point {
	return Point{p1[0] + p2[0], p1[1] + p2[1], p1[2] + p2[2]}
}

func (p1 *Point) Sub(p2 Point) Point {
	return Point{p1[0] - p2[0], p1[1] - p2[1], p1[2] - p2[2]}
}

func DS(p1, p2 *Point) int64 {
	x := p2.X() - p1.X()
	y := p2.Y() - p1.Y()
	z := p2.Z() - p1.Z()
	return x*x + y*y + z*z
}

func (p *Point) IsMainPoint() bool {
	allDivByThree := true
	for _, c := range *p {
		if c%THREE != 0 {
			allDivByThree = false
		}
	}
	return allDivByThree
}

func (p *Point) IsBorder(max int64) bool {
	for _, c := range *p {
		if c > 0 && c >= max-1 {
			return true
		}
		if c < 0 && c <= -max+1 {
			return true
		}
	}
	return false
}

func (s *Space) CreateStuff(max int64) {
	s.max = max
	org := s.createNodes()
	e := Event{0, org, 0, EventA, make([]*EventOutgrowth, 1, 100)}
	e.O[0] = &EventOutgrowth{org, &e, Distance(0)}
	org.E = make([]*EventOutgrowth, 1)
	org.E[0] = e.O[0]
	s.events[0] = &e
	s.createDrawingElements()
}

func (s *Space) ForwardTime() {
	s.current++
}

func (s *Space) BackTime() {
	s.current--
}

func (s *Space) GetNode(p *Point) *Node {
	for _, n := range s.nodes {
		if *(n.P) == *p {
			return n
		}
	}
	return nil
}

func (s *Space) getOrCreateNode(p *Point) *Node {
	n := s.GetNode(p)
	if n != nil {
		return n
	}
	n = &Node{}
	n.P = p
	s.nodes = append(s.nodes, n)
	if p.IsMainPoint() {
		s.createAndConnectBasePoints(n)
	}
	return n
}

func (s *Space) makeConnection(n1, n2 *Node) *Connection {
	if !n1.HasFreeConnections() {
		fmt.Println("Node 1", n1, "does not have free connections")
		return nil
	}
	if !n2.HasFreeConnections() {
		fmt.Println("Node 2", n2, "does not have free connections")
		return nil
	}
	if n2.P.IsMainPoint() {
		fmt.Println("Passing second point of connection", *(n2.P), "is a main point. Only P1 can be main")
		return nil
	}
	d := DS(n1.P, n2.P)
	if !(d == 2 || d == 3) {
		fmt.Println("Connection between 2 points", *(n1.P), *(n2.P), "that are not 2 or 3 DS away!")
		return nil
	}
	c := &Connection{n1, n2}
	s.connections = append(s.connections, c)
	n1done := false
	n2done := false
	for i := 0; i < THREE; i++ {
		if !n1done && n1.C[i] == nil {
			n1.C[i] = c
			n1done = true
		}
		if !n2done && n2.C[i] == nil {
			n2.C[i] = c
			n2done = true
		}
	}
	if !n1done || !n2done {
		fmt.Println("Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c
}

func (s *Space) createAndConnectBasePoints(n *Node) {
	if !n.P.IsMainPoint() {
		fmt.Println("Passing point to add base points", *(n.P), "is not a main point!")
		return
	}
	for _, b := range BasePoints {
		p2 := n.P.Add(b)
		bpn := s.getOrCreateNode(&p2)
		s.makeConnection(n, bpn)
	}
}

func (s *Space) createNodes() *Node {
	org := s.getOrCreateNode(&Origin)
	maxByThree := int64(s.max / THREE)
	fmt.Println("Max by three", maxByThree)
	for x := -maxByThree; x <= maxByThree; x++ {
		for y := -maxByThree; y <= maxByThree; y++ {
			for z := -maxByThree; z <= maxByThree; z++ {
				p := Point{x * THREE, y * THREE, z * THREE}
				s.getOrCreateNode(&p)
			}
		}
	}
	// All nodes that are not main with nil connections find good one
	for _, node := range s.nodes {
		if !node.P.IsMainPoint() && node.HasFreeConnections() {
			for _, other := range s.nodes {
				if node != other && !other.P.IsMainPoint() && other.HasFreeConnections() && DS(other.P, node.P) == 3 {
					s.makeConnection(node, other)
				}
				if !node.HasFreeConnections() {
					break
				}
			}
		}
	}

	// Verify all connections done
	for _, node := range s.nodes {
		for i, c := range node.C {
			// Should be on the border
			if c == nil && !node.P.IsBorder(s.max) {
				fmt.Println("something wrong with node connection not done for", node.P, "connection", i)
			}
		}
	}

	return org
}

func (n *Node) HasFreeConnections() bool {
	for _, c := range n.C {
		if c == nil {
			return true
		}
	}
	return false
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
		// For now just the first one
		if len(node.E) != 0 && node.E[0].E != nil {
			elements[offset] = &NodeDrawingElement{
				NodeActive, ObjectColor(1 + int8(node.E[0].E.K)), node,
			}
		} else {
			elements[offset] = &NodeDrawingElement{
				NodeEmpty, Grey, node,
			}
		}
		offset++
	}
	for _, conn := range s.connections {
		elements[offset] = MakeConnectionDrawingElement(conn.N1.P, conn.N2.P)
		offset++
	}
	fmt.Println("Created", len(elements), "elements.")
	s.Elements = elements
}

type SpaceDrawingElement interface {
	Key() ObjectType
	Color() ObjectColor
	Pos() *Point
}

type NodeDrawingElement struct {
	t ObjectType
	c ObjectColor
	n *Node
}

type ConnectionDrawingElement struct {
	t      ObjectType
	c      ObjectColor
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

func (n *NodeDrawingElement) Color() ObjectColor {
	return n.c
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
				return &ConnectionDrawingElement{ObjectType(int(Connection1) + i), Grey, p1, p2,}
			}
		}
		fmt.Println("What 1", p1, p2, bv)
		return &ConnectionDrawingElement{Connection1, Grey,p1, p2,}
	} else if p2.IsMainPoint() {
		bv = bv.Mul(-1)
		for i, bp := range BasePoints {
			if bp[0] == bv[0] && bp[1] == bv[1] && bp[2] == bv[2] {
				return &ConnectionDrawingElement{ObjectType(int(Connection1) + i), Grey, p2, p1,}
			}
		}
		fmt.Println("What 2", p1, p2, bv)
		return &ConnectionDrawingElement{Connection1, Grey, p2, p1,}
	} else {
		if bv[0] == 1 {
			if bv[1] != -1 || bv[2] != -1 {
				fmt.Println("What 3", p1, p2, bv)
			}
			return &ConnectionDrawingElement{Connection4, Grey, p1, p2,}
		} else {
			if bv[0] != -1 || bv[1] != 1 || bv[2] != 1 {
				fmt.Println("What 4", p1, p2, bv)
			}
			return &ConnectionDrawingElement{Connection5, Grey, p1, p2,}
		}
	}
}

func (c *ConnectionDrawingElement) Key() ObjectType {
	return c.t
}

func (c *ConnectionDrawingElement) Color() ObjectColor {
	return c.c
}

func (c *ConnectionDrawingElement) Pos() *Point {
	return c.p1
}

// AxeDrawingElement functions
func (a *AxeDrawingElement) Key() ObjectType {
	return a.t
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
