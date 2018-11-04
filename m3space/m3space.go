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

func (space *Space) Clear() {
	space.events = make(map[EventID]*Event)
	space.nodesMap = make(map[Point]*Node)
	space.connections = make([]*Connection, 0, 500)
	space.currentId = 0
	space.currentTime = 0
	space.max = 9
	space.Elements = make([]SpaceDrawingElement, 0, 500)
}

func (space *Space) SetMax(max int64) {
	space.max = max
}

func (space *Space) CreateSingleEventCenter() {
	space.CreateEvent(Origin, GreenEvent)
	space.createDrawingElements()
}

func (space *Space) CreatePyramid(pyramidSize int64) {
	space.CreateEvent(Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	space.CreateEvent(Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	space.CreateEvent(Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	space.CreateEvent(Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
	space.createDrawingElements()
}

func (space *Space) ForwardTime() {
	for _, evt := range space.events {
		evt.createNewOutgrowths()
	}
	// Switch latest to old, and new to latest
	for _, evt := range space.events {
		evt.moveNewOutgrowthsToLatest()
	}
	space.currentTime++
	// Same drawing elements just changed color :(
	space.createDrawingElements()
}

func (space *Space) BackTime() {
	fmt.Println("Very hard to go back in time !!!")
	//space.currentTime--
}

func (space *Space) GetNode(p Point) *Node {
	n, ok := space.nodesMap[p]
	if ok {
		return n
	}
	return nil
}

func (space *Space) getOrCreateNode(p Point) *Node {
	n := space.GetNode(p)
	if n != nil {
		return n
	}
	n = &Node{&p, nil, nil, }
	space.nodesMap[p] = n
	return n
}

func (space *Space) makeConnection(n1, n2 *Node) *Connection {
	if !n1.HasFreeConnections() {
		fmt.Println("Node 1", n1, "does not have free connections")
		return nil
	}
	if !n2.HasFreeConnections() {
		fmt.Println("Node 2", n2, "does not have free connections")
		return nil
	}
	// Flipping if needed to make sure n1 is main
	if n2.point.IsMainPoint() {
		temp := n1
		n1 = n2
		n2 = temp
	}
	d := DS(n1.point, n2.point)
	if !(d == 1 || d == 2 || d == 3 || d == 5) {
		fmt.Println("Connection between 2 points", *(n1.point), *(n2.point), "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// Verify not already connected
	if n1.IsAlreadyConnected(n2) {
		fmt.Println("Connection between 2 points", *(n1.point), *(n2.point), "already connected!")
		return nil
	}
	// All good create connection
	c := &Connection{n1, n2}
	space.connections = append(space.connections, c)
	n1done := n1.AddConnection(c)
	n2done := n2.AddConnection(c)
	if n1done < 0 || n2done < 0 {
		fmt.Println("Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c
}

func (space *Space) createDrawingElements() {
	nbElements := 6 + len(space.nodesMap) + len(space.connections)
	elements := make([]SpaceDrawingElement, nbElements)
	offset := 0
	for axe := 0; axe < 3; axe++ {
		elements[offset] = &AxeDrawingElement{
			ObjectType(axe),
			space.max + AxeExtraLength,
			false,
		}
		offset++
		elements[offset] = &AxeDrawingElement{
			ObjectType(axe),
			space.max + AxeExtraLength,
			true,
		}
		offset++
	}
	for _, node := range space.nodesMap {
		elements[offset] = MakeNodeDrawingElement(node)
		offset++
	}
	for _, conn := range space.connections {
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
	space.Elements = elements
}

func (space *Space) DisplaySettings() {
	fmt.Println("========= Space Settings =========")
	fmt.Println("Current Time", space.currentTime)
	fmt.Println("Nb Nodes", len(space.nodesMap), ", Nb Connections", len(space.connections), ", Nb Events", len(space.events))
}
