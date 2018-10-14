package m3space

import (
	"fmt"
)

const (
	AxeExtraLength = 3
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

var DEBUG = false

type TickTime uint64

type Distance uint64

type EventID uint64

type EventColor int8

type EventOutgrowthState int8

const (
	EventA EventColor = iota
	EventB
	EventC
	EventD
)

type ObjectColor int8

const (
	Grey   ObjectColor = iota
	Red
	Green
	Blue
	Yellow
)

const (
	EventOutgrowthLatest EventOutgrowthState = iota
	EventOutgrowthNew
	EventOutgrowthOld
	EventOutgrowthDead
)

type Event struct {
	ID EventID
	N  *Node
	T  TickTime
	K  EventColor
	O  []*EventOutgrowth
}

type EventOutgrowth struct {
	N *Node
	E *Event
	D Distance
	S EventOutgrowthState
}

type Space struct {
	nodes       []*Node
	connections []*Connection
	currentId   EventID
	events      map[EventID]*Event
	currentTime TickTime
	max         int64
	Elements    []SpaceDrawingElement
}

var SpaceObj = Space{
	make([]*Node, 0, 108),
	make([]*Connection, 0, 500),
	0,
	make(map[EventID]*Event),
	0,
	9,
	make([]SpaceDrawingElement, 0, 500),
}

func (s *Space) CreateStuff(max int64) {
	s.max = max
	s.createNodes()
	pyramidSize := int64(s.max/THREE) / 2 + 1
	s.CreateEvent(Point{3, 0, 3}.Mul(pyramidSize), EventA)
	s.CreateEvent(Point{-3, 3, 3}.Mul(pyramidSize), EventB)
	s.CreateEvent(Point{-3, -3, 3}.Mul(pyramidSize), EventC)
	s.CreateEvent(Point{0, 0, -3}.Mul(pyramidSize), EventD)
	s.createDrawingElements()
}

func (s *Space) ForwardTime() {
	for _, evt := range s.events {
		for _, eg := range evt.O {
			if eg.S == EventOutgrowthLatest {
				for _, c := range eg.N.C {
					if c != nil {
						otherNode := c.N1
						if otherNode == eg.N {
							otherNode = c.N2
						}
						hasAlreadyEvent := false
						for _, eo := range otherNode.E {
							if eo.E.ID == evt.ID {
								hasAlreadyEvent = true
							}
						}
						if !hasAlreadyEvent {
							if DEBUG {
								fmt.Println("Creating new event outgrowth for", evt.ID, "at", otherNode.P)
							}
							newEo := &EventOutgrowth{otherNode, evt, eg.D + 1, EventOutgrowthNew}
							otherNode.E = append(otherNode.E, newEo)
							evt.O = append(evt.O, newEo)
						}
					}
				}
			}
		}
	}
	// Switch latest to old, and new to latest
	for _, evt := range s.events {
		for _, eg := range evt.O {
			switch eg.S {
			case EventOutgrowthLatest:
				eg.S = EventOutgrowthOld
			case EventOutgrowthNew:
				eg.S = EventOutgrowthLatest
			}
		}
	}
	s.currentTime++
	// Same drawing elements just changed color :(
	s.createDrawingElements()
}

func (s *Space) BackTime() {
	s.currentTime--
}

func (s *Space) CreateEvent(p Point, k EventColor) *Event {
	n := s.GetNode(&p)
	if n == nil {
		fmt.Println("Creating event on non existent node, on point", p, "kind", k)
		return nil
	}
	id := s.currentId
	s.currentId++
	e := Event{id, n, s.currentTime, k, make([]*EventOutgrowth, 1, 100)}
	e.O[0] = &EventOutgrowth{n, &e, Distance(0), EventOutgrowthLatest}
	n.E = make([]*EventOutgrowth, 1)
	n.E[0] = e.O[0]
	s.events[id] = &e
	return &e
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
	if DEBUG {
		fmt.Println("Max by three", maxByThree)
	}
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

func (s *Space) createDrawingElements() {
	nbElements := 6+len(s.nodes)+len(s.connections)
	elements := make([]SpaceDrawingElement, nbElements)
	offset := 0
	for axe := 0; axe < 3; axe++ {
		elements[offset] = &AxeDrawingElement{
			ObjectType(axe),
			s.max + AxeExtraLength,
			false,
		}
		offset++
		elements[offset] = &AxeDrawingElement{
			ObjectType(axe),
			s.max + AxeExtraLength,
			true,
		}
		offset++
	}
	for _, node := range s.nodes {
		// For now just the first one
		if len(node.E) != 0 && node.E[0].E != nil {
			a := float32(0.2)
			if node.E[0].S == EventOutgrowthLatest {
				a = float32(1.0)
			}
			elements[offset] = &NodeDrawingElement{
				NodeActive, ObjectColor(1 + int8(node.E[0].E.K)), a, node,
			}
		} else {
			elements[offset] = &NodeDrawingElement{
				NodeEmpty, Grey, 0.7, node,
			}
		}
		offset++
	}
	for _, conn := range s.connections {
		elements[offset] = MakeConnectionDrawingElement(conn.N1.P, conn.N2.P)
		offset++
	}
	if offset != nbElements {
		fmt.Println("Created", offset, "elements, but it should be",nbElements)
		return
	}
	if DEBUG {
		fmt.Println("Created", nbElements, "elements.")
	}
	s.Elements = elements
}

