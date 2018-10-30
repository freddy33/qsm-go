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

type EventColor uint8

type EventOutgrowthState uint8

const (
	RedEvent    EventColor = iota
	GreenEvent
	BlueEvent
	YellowEvent
)

const (
	EventOutgrowthLatest EventOutgrowthState = iota
	EventOutgrowthNew
	EventOutgrowthOld
	EventOutgrowthDead
)

type Event struct {
	id            EventID
	node          *Node
	created       TickTime
	color         EventColor
	growthContext GrowthContext
	outgrowths    []*EventOutgrowth
	newOutgrowths []*EventOutgrowth
}

type EventOutgrowth struct {
	node     *Node
	event    *Event
	from     *EventOutgrowth
	distance Distance
	state    EventOutgrowthState
}

type Space struct {
	events      map[EventID]*Event
	nodesMap    map[Point]*Node
	connections []*Connection
	currentId   EventID
	currentTime TickTime
	max         int64
	Elements    []SpaceDrawingElement
}

var SpaceObj = Space{}

func init() {
	SpaceObj.Clear()
}

func (s *Space) Clear() {
	s.events = make(map[EventID]*Event)
	s.nodesMap = make(map[Point]*Node)
	s.connections = make([]*Connection, 0, 500)
	s.currentId = 0
	s.currentTime = 0
	s.max = 9
	s.Elements = make([]SpaceDrawingElement, 0, 500)
}

func (s *Space) SetMax(max int64) {
	s.max = max
}

func (s *Space) CreateSingleEventCenter() {
	s.CreateEvent(Origin, GreenEvent)
	s.createDrawingElements()
}

func (s *Space) CreatePyramid(pyramidSize int64) {
	s.CreateEvent(Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	s.CreateEvent(Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	s.CreateEvent(Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	s.CreateEvent(Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
	s.createDrawingElements()
}

func (evt *Event) createNewOutgrowths(s *Space) {
	evt.newOutgrowths = evt.newOutgrowths[:0]
	for _, eg := range evt.outgrowths {
		if eg.state == EventOutgrowthLatest {
			nextPoints := eg.node.point.getNextPoints(&(evt.growthContext))
			connToPoint := [3]int{-1,-1,-1}

			for connIdx, conn := range eg.node.connections {
				if conn != nil {
					for pointIdx, nextPoint := range nextPoints {
						if nextPoint == *(conn.N1.point) || nextPoint == *(conn.N2.point) {
							connToPoint[connIdx] = pointIdx
						}
					}
				}
			}

			if eg.node.HasFreeConnections() {
				for pointIdx, _ := range nextPoints {
					nextPoint := nextPoints[pointIdx]
					alreadyMapped := false
					for _, cIdx := range connToPoint {
						if cIdx == pointIdx {
							alreadyMapped = true
						}
					}
					if !alreadyMapped {
						otherNode := s.getOrCreateNode(&nextPoint)
						s.makeConnection(eg.node, otherNode)
						for connIdx, conn := range eg.node.connections {
							if conn != nil && (nextPoint == *(conn.N1.point) || nextPoint == *(conn.N2.point)) {
								connToPoint[connIdx] = pointIdx
							}
						}
					}
					if !eg.node.HasFreeConnections() {
						break
					}
				}
			}

			for connIdx, pointIdx := range connToPoint {
				if pointIdx != -1 {
					c := eg.node.connections[connIdx]
					otherNode := c.N1
					if otherNode == eg.node {
						otherNode = c.N2
					}
					// Roots cannot have outgrowth that not theirs (TODO: why?)
					hasAlreadyEvent := otherNode.IsRoot()
					for _, eo := range otherNode.outgrowths {
						if eo.event.id == evt.id {
							hasAlreadyEvent = true
						}
					}
					if !hasAlreadyEvent {
						if DEBUG {
							fmt.Println("Creating new event outgrowth for", evt.id, "at", otherNode.point)
						}
						newEo := &EventOutgrowth{otherNode, evt, eg, eg.distance + 1, EventOutgrowthNew}
						otherNode.outgrowths = append(otherNode.outgrowths, newEo)
						evt.newOutgrowths = append(evt.newOutgrowths, newEo)
					}
				}
			}
		}
	}
}

func (evt *Event) moveNewOutgrowthsToLatest() {
	for _, eg := range evt.outgrowths {
		switch eg.state {
		case EventOutgrowthLatest:
			eg.state = EventOutgrowthOld
		}
	}
	for _, eg := range evt.newOutgrowths {
		switch eg.state {
		case EventOutgrowthNew:
			eg.state = EventOutgrowthLatest
			evt.outgrowths = append(evt.outgrowths, eg)
		}
	}
}

func (s *Space) ForwardTime() {
	for _, evt := range s.events {
		evt.createNewOutgrowths(s)
	}
	// Switch latest to old, and new to latest
	for _, evt := range s.events {
		evt.moveNewOutgrowthsToLatest()
	}
	s.currentTime++
	// Same drawing elements just changed color :(
	s.createDrawingElements()
}

func (s *Space) BackTime() {
	fmt.Println("Very hard to go back in time !!!")
	//s.currentTime--
}

func (s *Space) CreateEvent(p Point, k EventColor) *Event {
	n := s.getOrCreateNode(&p)
	id := s.currentId
	s.currentId++
	ctx := GrowthContext{8, 0, false, 0}
	switch k {
	case RedEvent:
		// No change
	case GreenEvent:
		ctx.permutationIndex = 3
		ctx.permutationOffset = 2
	case BlueEvent:
		ctx.permutationIndex = 6
		ctx.permutationOffset = 4
	case YellowEvent:
		ctx.permutationIndex = 9
		ctx.permutationOffset = 6
	}
	e := Event{id, n, s.currentTime, k,
	ctx,
	make([]*EventOutgrowth, 1, 100), make([]*EventOutgrowth, 0, 10),}
	e.outgrowths[0] = &EventOutgrowth{n, &e, nil, Distance(0), EventOutgrowthLatest}
	n.outgrowths = make([]*EventOutgrowth, 1)
	n.outgrowths[0] = e.outgrowths[0]
	s.events[id] = &e
	return &e
}

func (s *Space) GetNode(p *Point) *Node {
	n, ok := s.nodesMap[*p]
	if ok {
		return n
	}
	return nil
}

func (s *Space) getOrCreateNode(p *Point) *Node {
	n := s.GetNode(p)
	if n != nil {
		return n
	}
	n = &Node{}
	n.point = p
	s.nodesMap[*p] = n
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
	if n2.point.IsMainPoint() {
		fmt.Println("Passing second point of connection", *(n2.point), "is a main point. Only P1 can be main")
		return nil
	}
	d := DS(n1.point, n2.point)
	if !(d == 1 || d == 2 || d == 3 || d == 5) {
		fmt.Println("Connection between 2 points", *(n1.point), *(n2.point), "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// Verify not already connected
	for i := 0; i < THREE; i++ {
		if n1.connections[i] != nil && (n1.connections[i].N1 == n2 || n1.connections[i].N2 == n2) {
			if DEBUG {
				fmt.Println("Connection between 2 points", *(n1.point), *(n2.point), "already connected!")
			}
			return nil
		}
		if n2.connections[i] != nil && (n2.connections[i].N1 == n1 || n2.connections[i].N2 == n1) {
			if DEBUG {
				fmt.Println("Connection between 2 points", *(n1.point), *(n2.point), "already connected!")
			}
			return nil
		}
	}

	// All good create connection
	c := &Connection{n1, n2}
	s.connections = append(s.connections, c)
	n1done := false
	n2done := false
	for i := 0; i < THREE; i++ {
		if !n1done && n1.connections[i] == nil {
			n1.connections[i] = c
			n1done = true
		}
		if !n2done && n2.connections[i] == nil {
			n2.connections[i] = c
			n2done = true
		}
	}
	if !n1done || !n2done {
		fmt.Println("Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c
}

func (s *Space) createDrawingElements() {
	nbElements := 6 + len(s.nodesMap) + len(s.connections)
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
	for _, node := range s.nodesMap {
		elements[offset] = MakeNodeDrawingElement(node)
		offset++
	}
	for _, conn := range s.connections {
		elements[offset] = MakeConnectionDrawingElement(conn)
		offset++
	}
	if offset != nbElements {
		fmt.Println("Created", offset, "elements, but it should be", nbElements)
		return
	}
	if DEBUG {
		fmt.Println("Created", nbElements, "elements.")
	}
	s.Elements = elements
}

func (s *Space) DisplaySettings() {
	fmt.Println("========= Space Settings =========")
	fmt.Println("Current Time", s.currentTime)
	fmt.Println("Nb Nodes", len(s.nodesMap), ", Nb Connections", len(s.connections), ", Nb Events", len(s.events))
}
