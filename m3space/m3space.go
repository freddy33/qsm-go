package m3space

import (
	"fmt"
)

const (
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

var DEBUG = false

type TickTime uint64

type SpaceVisitor interface {
	VisitNode(space *Space, node *Node)
	VisitConnection(space *Space, conn *Connection)
}

type Space struct {
	events      map[EventID]*Event
	nodesMap    map[Point]*Node
	connections []*Connection
	currentId   EventID
	currentTime TickTime
	// Max size of all CurrentSpace. TODO: Make it variable using the furthest node from origin
	Max int64
	// Max number of connections per node
	MaxConnections int
	// Distance from latest to consider event outgrowth active
	EventOutgrowthThreshold Distance
}

func MakeSpace(max int64) Space {
	space := Space{}
	space.events = make(map[EventID]*Event)
	space.nodesMap = make(map[Point]*Node)
	space.connections = make([]*Connection, 0, 500)
	space.currentId = 0
	space.currentTime = 0
	space.Max = max
	space.MaxConnections = 3
	space.EventOutgrowthThreshold = Distance(1)
	return space
}

func (space *Space) GetCurrentTime() TickTime {
	return space.currentTime
}

func (space *Space) GetNbNodes() int {
	return len(space.nodesMap)
}

func (space *Space) GetNbConnections() int {
	return len(space.connections)
}

func (space *Space) GetNbEvents() int {
	return len(space.events)
}

func (space *Space) VisitAll(visitor SpaceVisitor) {
	for _, node := range space.nodesMap {
		visitor.VisitNode(space, node)
	}
	for _, conn := range space.connections {
		visitor.VisitConnection(space, conn)
	}
}

func (space *Space) CreateSingleEventCenter() *Event {
	return space.CreateEvent(Origin, RedEvent)
}

func (space *Space) CreatePyramid(pyramidSize int64) {
	space.CreateEvent(Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	space.CreateEvent(Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	space.CreateEvent(Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	space.CreateEvent(Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
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
	n = &Node{&p, nil, nil,}
	space.nodesMap[p] = n
	return n
}

func (space *Space) makeConnection(n1, n2 *Node) *Connection {
	if !n1.HasFreeConnections(space) {
		if DEBUG {
			fmt.Println("Node 1", n1, "does not have free connections")
		}
		return nil
	}
	if !n2.HasFreeConnections(space) {
		if DEBUG {
			fmt.Println("Node 2", n2, "does not have free connections")
		}
		return nil
	}
	if n1.IsAlreadyConnected(n2) {
		if DEBUG {
			fmt.Println("Connection between 2 points", *(n1.Pos), *(n2.Pos), "already connected!")
		}
		return nil
	}

	// Flipping if needed to make sure n1 is main
	if n2.Pos.IsMainPoint() {
		temp := n1
		n1 = n2
		n2 = temp
	}
	d := DS(n1.Pos, n2.Pos)
	if !(d == 1 || d == 2 || d == 3 || d == 5) {
		fmt.Println("ERROR: Connection between 2 points", *(n1.Pos), *(n2.Pos), "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// All good create connection
	c := &Connection{n1, n2}
	space.connections = append(space.connections, c)
	n1done := n1.AddConnection(c, space)
	n2done := n2.AddConnection(c, space)
	if n1done < 0 || n2done < 0 {
		fmt.Println("ERROR: Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c
}

func (space *Space) DisplaySettings() {
	fmt.Println("========= Space Settings =========")
	fmt.Println("Current Time", space.currentTime)
	fmt.Println("Nb Nodes", len(space.nodesMap), ", Nb Connections", len(space.connections), ", Nb Events", len(space.events))
}
