package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
)

var Log = m3util.NewLogger("m3space", m3util.INFO)

const (
	// Where the number matters and appear. Remember that 3 is the number!
	THREE = 3
)

type TickTime uint64

type SpaceVisitor interface {
	VisitNode(space *Space, node *ActiveNode)
	VisitConnection(space *Space, conn *Connection)
}

type Space struct {
	events            map[EventID]*Event
	activeNodesMap    map[Point]*ActiveNode
	oldNodesMap       map[Point]*SavedNode
	activeConnections []*Connection
	nbOldConnections  int
	currentId         EventID
	currentTime       TickTime
	// Max absolute coordinate in all nodes
	Max int64
	// Max number of connections per node
	MaxConnections int
	// Cancel on same event conflict
	blockOnSameEvent int
	// Distance from latest below which to consider event outgrowth active
	EventOutgrowthThreshold Distance
	// Distance from latest above which to consider event outgrowth old
	EventOutgrowthOldThreshold Distance
}

func MakeSpace(max int64) Space {
	space := Space{}
	space.events = make(map[EventID]*Event)
	space.activeNodesMap = make(map[Point]*ActiveNode)
	space.oldNodesMap = make(map[Point]*SavedNode)
	space.activeConnections = make([]*Connection, 0, 500)
	space.nbOldConnections = 0
	space.currentId = 0
	space.currentTime = 0
	space.Max = max
	space.MaxConnections = 3
	space.blockOnSameEvent = 3
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
	return len(space.activeConnections) + space.nbOldConnections
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

func (space *Space) GetNode(p Point) Node {
	n, ok := space.activeNodesMap[p]
	if ok {
		return n
	}
	sn, ok := space.oldNodesMap[p]
	if ok {
		return sn
	}
	return nil
}

func (space *Space) getAndActivateNode(p Point) *ActiveNode {
	n, ok := space.activeNodesMap[p]
	if ok {
		return n
	}
	sn, ok := space.oldNodesMap[p]
	if ok {
		Log.Debugf("Recovering node %s from storage to active", sn.GetStateString())
		// becomes active
		delete(space.oldNodesMap, p)
		n = NewNode(p)
		n.root = sn.IsRoot()
		n.accessedEventIDS = sn.accessedEventIDS
		n.connections = make([]*Connection, len(sn.connections))
		for i, connId := range sn.connections {
			cd := AllConnectionsIds[connId]
			p2 := n.Pos.Add(cd.Vector)
			n.connections[i] = &Connection{connId, n.Pos, p2,}
		}
		space.activeNodesMap[p] = n
		return n
	}
	return nil
}

func (space *Space) getOrCreateNode(p Point) *ActiveNode {
	n := space.getAndActivateNode(p)
	if n != nil {
		return n
	}
	n = NewNode(p)
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

func (space *Space) DisplayState() {
	fmt.Println("========= Space State =========")
	fmt.Println("Current Time", space.currentTime)
	fmt.Println("Nb Active Nodes", len(space.activeNodesMap), "Nb Old Nodes", len(space.oldNodesMap), ", Nb Connections", len(space.activeConnections), ", Nb Events", len(space.events))
}
