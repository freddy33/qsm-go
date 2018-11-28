package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
)

const (
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

var Log = m3util.NewLogger("m3space", m3util.INFO)

type TickTime uint64

type SpaceVisitor interface {
	VisitNode(space *Space, node *Node)
	VisitConnection(space *Space, conn *Connection)
}

type Space struct {
	events            map[EventID]*Event
	activeNodesMap    map[Point]*Node
	oldNodesMap       map[Point]*Node
	activeConnections []*Connection
	oldConnections []*Connection
	currentId         EventID
	currentTime       TickTime
	// Max size of all CurrentSpace. TODO: Make it variable using the furthest node from origin
	Max int64
	// Max number of connections per node
	MaxConnections int
	// Distance from latest below which to consider event outgrowth active
	EventOutgrowthThreshold Distance
	// Distance from latest above which to consider event outgrowth old
	EventOutgrowthOldThreshold Distance
}

func MakeSpace(max int64) Space {
	space := Space{}
	space.events = make(map[EventID]*Event)
	space.activeNodesMap = make(map[Point]*Node)
	space.oldNodesMap = make(map[Point]*Node)
	space.activeConnections = make([]*Connection, 0, 500)
	space.currentId = 0
	space.currentTime = 0
	space.Max = max
	space.MaxConnections = 3
	space.SetEventOutgrowthThreshold(Distance(1))
	return space
}

func (space *Space) SetEventOutgrowthThreshold(threshold Distance) {
	if threshold > 2^50 {
		threshold = 0
	}
	space.EventOutgrowthThreshold = threshold
	// Everything more than 3 above threshold move to dead => old
	space.EventOutgrowthOldThreshold = threshold + 3
}

func (space *Space) GetCurrentTime() TickTime {
	return space.currentTime
}

func (space *Space) GetNbActiveNodes() int {
	return len(space.activeNodesMap)
}

func (space *Space) GetNbNodes() int {
	return len(space.activeNodesMap) + len(space.oldNodesMap)
}

func (space *Space) GetNbActiveConnections() int {
	return len(space.activeConnections)
}

func (space *Space) GetNbConnections() int {
	return len(space.activeConnections) + len(space.oldConnections)
}

func (space *Space) GetNbEvents() int {
	return len(space.events)
}

func (space *Space) VisitAll(visitor SpaceVisitor) {
	for _, node := range space.activeNodesMap {
		visitor.VisitNode(space, node)
	}
	for _, conn := range space.activeConnections {
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
	n, ok := space.activeNodesMap[p]
	if ok {
		return n
	}
	n, ok = space.oldNodesMap[p]
	if ok {
		return n
	}
	return nil
}

func (space *Space) getAndActivateNode(p Point) *Node {
	n, ok := space.activeNodesMap[p]
	if ok {
		return n
	}
	n, ok = space.oldNodesMap[p]
	if ok {
		if !n.IsOld(space) {
			// becomes active
			delete(space.oldNodesMap, p)
			space.activeNodesMap[p] = n
		}
		return n
	}
	return nil
}

func (space *Space) getOrCreateNode(p Point) *Node {
	n := space.getAndActivateNode(p)
	if n != nil {
		return n
	}
	n = NewNode(&p)
	space.activeNodesMap[p] = n
	for _, c := range p {
		if c > 0 && space.Max < c {
			space.Max = c
		}
		if c < 0 && space.Max < -c {
			space.Max = -c
		}
	}
	return n
}

func (space *Space) makeConnection(n1, n2 *Node) *Connection {
	if !n1.HasFreeConnections(space) {
		Log.Trace("Node 1", n1, "does not have free connections")
		return nil
	}
	if !n2.HasFreeConnections(space) {
		Log.Trace("Node 2", n2, "does not have free connections")
		return nil
	}
	if n1.IsAlreadyConnected(n2) {
		Log.Trace("Connection between 2 points", *(n1.Pos), *(n2.Pos), "already connected!")
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
		Log.Error("Connection between 2 points", *(n1.Pos), *(n2.Pos), "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// All good create connection
	c := &Connection{n1, n2}
	space.activeConnections = append(space.activeConnections, c)
	n1done := n1.AddConnection(c, space)
	n2done := n2.AddConnection(c, space)
	if n1done < 0 || n2done < 0 {
		Log.Error("Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c
}

func (space *Space) DisplayState() {
	fmt.Println("========= Space State =========")
	fmt.Println("Current Time", space.currentTime)
	fmt.Println("Nb Active Nodes", len(space.activeNodesMap),"Nb Old Nodes", len(space.oldNodesMap), ", Nb Connections", len(space.activeConnections), ", Nb Events", len(space.events))
}
