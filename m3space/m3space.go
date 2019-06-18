package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
)

var Log = m3util.NewLogger("m3space", m3util.INFO)

type SpaceVisitor interface {
	VisitNode(space *Space, node Node)
	VisitConnection(space *Space, conn m3path.PathLink)
}

type Space struct {
	// the int value of the next event id created
	lastIdCounter EventID

	// The slice of events where the index is the EventID
	events []*Event

	// The current time of space time
	currentTime DistAndTime

	// The single big map of all the points
	nodesMap map[m3point.Point]Node
	// Extracted list from the above map of the current state at currentTime
	activeNodes []Node
	activeLinks []m3path.PathLink

	nbDeadNodes int

	// Max absolute coordinate in all nodes
	Max int64
	// Max number of connections per node
	MaxConnections int
	// Cancel on same event conflict
	blockOnSameEvent int
	// DistAndTime from latest below which to consider event outgrowth active
	EventOutgrowthThreshold DistAndTime
	// DistAndTime from latest above which to consider event outgrowth old
	EventOutgrowthOldThreshold DistAndTime
	// DistAndTime from latest above which to consider event outgrowth dead
	EventOutgrowthDeadThreshold DistAndTime
}

func MakeSpace(max int64) Space {
	space := Space{}
	space.lastIdCounter = 1
	space.events = make([]*Event, 0, 12)
	space.currentTime = 0
	space.nodesMap = make(map[m3point.Point]Node, 1000)
	space.activeNodes = make([]Node, 0, 500)
	space.activeLinks = make([]m3path.PathLink, 0, 500)

	space.nbDeadNodes = 0
	space.Max = max
	space.MaxConnections = 3
	space.blockOnSameEvent = 3
	space.SetEventOutgrowthThreshold(DistAndTime(1))
	return space
}

func (space *Space) SetEventOutgrowthThreshold(threshold DistAndTime) {
	if threshold > 2^50 {
		threshold = 0
	}
	space.EventOutgrowthThreshold = threshold
	// Everything more than 3*3 above threshold move to active => old
	space.EventOutgrowthOldThreshold = threshold + 3
	// Everything more than 3*3*3 above threshold move to old => dead
	space.EventOutgrowthDeadThreshold = threshold + 3*3
}

func (space *Space) GetCurrentTime() DistAndTime {
	return space.currentTime
}

func (space *Space) GetNbActiveNodes() int {
	return len(space.activeNodes)
}

func (space *Space) GetNbNodes() int {
	return len(space.nodesMap)
}

func (space *Space) GetNbActiveLinks() int {
	return len(space.activeLinks)
}

func (space *Space) GetNbEvents() int {
	return len(space.events)
}

func (space *Space) GetEvent(id EventID) *Event {
	return space.events[id]
}

func (space *Space) VisitAll(visitor SpaceVisitor) {
	for _, n := range space.activeNodes {
		visitor.VisitNode(space, n)
	}
	for _, pl := range space.activeLinks {
		visitor.VisitConnection(space, pl)
	}
}

func (space *Space) CreateSingleEventCenter() *Event {
	return space.CreateEventFromColor(m3point.Origin, RedEvent)
}

func (space *Space) CreatePyramid(pyramidSize int64) {
	space.CreateEventFromColor(m3point.Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	space.CreateEventFromColor(m3point.Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	space.CreateEventFromColor(m3point.Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	space.CreateEventFromColor(m3point.Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
}

func (space *Space) GetNode(p m3point.Point) Node {
	return space.nodesMap[p]
}

func (space *Space) newEmptyNode() Node {
	an := new(PointNode)
	an.pathNodes = make([]m3path.PathNode, space.lastIdCounter)
	return an
}

func (space *Space) getOrCreateNode(p m3point.Point) Node {
	n := space.GetNode(p)
	if n != nil {
		return n
	}
	n = space.newEmptyNode()
	space.nodesMap[p] = n
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
	fmt.Println("Nb Nodes", len(space.nodesMap), "Nb Active Nodes", len(space.activeNodes), ", Nb Connections", len(space.activeLinks), ", Nb Events", len(space.events))
}
