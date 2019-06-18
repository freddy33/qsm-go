package m3space

import (
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
)

type EventID int

const (
	NilEvent = EventID(-1)
)

type DistAndTime int

type EventColor uint8

const (
	RedEvent EventColor = 1 << iota
	GreenEvent
	BlueEvent
	YellowEvent
)

var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type Event struct {
	id          EventID
	space       *Space
	node        Node
	created     DistAndTime
	color       EventColor
	pathContext *m3path.PathContext
}

type SpacePathNodeMap struct {
	space *Space
	id EventID
	size int
}

/***************************************************************/
// SpacePathNodeMap Functions
/***************************************************************/

func (spnm *SpacePathNodeMap) GetSize() int {
	return spnm.size
}

func (spnm *SpacePathNodeMap) GetPathNode(p m3point.Point) (m3path.PathNode, bool) {
	res, ok := spnm.space.nodesMap[p]
	if ok {
		return res.GetPathNode(spnm.id), ok
	}
	return nil, false
}

func (spnm *SpacePathNodeMap) AddPathNode(pathNode m3path.PathNode) {
	n := spnm.space.getOrCreateNode(pathNode.P())
	n.addPathNode(spnm.id, pathNode)
	spnm.size++
	spnm.space.activeNodes = append(spnm.space.activeNodes, n)
}

/***************************************************************/
// Event Functions
/***************************************************************/

func (space *Space) CreateEvent(ctxType m3point.ContextType, idx int, offset int, p m3point.Point, k EventColor) *Event {
	pnm := &SpacePathNodeMap{space, space.lastIdCounter, 0}
	space.lastIdCounter++
	ctx := m3path.MakePathContext(ctxType, idx, offset, pnm)
	e := Event{pnm.id, space, nil, space.currentTime, k, ctx}
	space.events[pnm.id] = &e
	ctx.InitRootNode(p)
	e.node = space.GetNode(p)
	return &e
}

func (space *Space) CreateEventFromColor(p m3point.Point, k EventColor) *Event {
	idx, offset := getIndexAndOffsetForColor(k)
	return space.CreateEvent(8, idx, offset, p, k)
}

func getIndexAndOffsetForColor(k EventColor) (int, int) {
	switch k {
	case RedEvent:
		return 0, 0
	case GreenEvent:
		return 4, 0
	case BlueEvent:
		return 8, 0
	case YellowEvent:
		return 10, 4
	}
	Log.Errorf("Event color unknown %v", k)
	return -1, -1
}

func (evt *Event) LatestDistance() DistAndTime {
	// DistAndTime and time are the same...
	return DistAndTime(evt.space.currentTime - evt.created)
}
