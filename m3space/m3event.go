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

// TODO: This should be in the space data entry of the environment
var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type Event struct {
	id          EventID
	space       *Space
	pathNodeMap m3path.PathNodeMap
	node        Node
	created     DistAndTime
	color       EventColor
	pathContext m3path.PathContext
}

type SpacePathNodeMap struct {
	space *Space
	id EventID
	size int
}

func (spnm *SpacePathNodeMap) Size() int {
	panic("implement me")
}

func (spnm *SpacePathNodeMap) Clear() {
	panic("implement me")
}

func (spnm *SpacePathNodeMap) Range(f func(point m3point.Point, pn m3path.PathNode) bool, nbProc int) {
	panic("implement me")
}

/***************************************************************/
// SpacePathNodeMap Functions
/***************************************************************/

func (spnm *SpacePathNodeMap) GetSize() int {
	return spnm.size
}

func (spnm *SpacePathNodeMap) GetPathNode(p m3point.Point) m3path.PathNode {
	res, ok := spnm.space.nodesMap.Load(p)
	if ok {
		pathNode := res.(Node).GetPathNode(spnm.id)
		if pathNode != nil {
			return pathNode
		}
	}
	return nil
}

func (spnm *SpacePathNodeMap) AddPathNode(pathNode m3path.PathNode) (m3path.PathNode, bool) {
	n := spnm.space.getOrCreateNode(pathNode.P())
	nbLatest := n.GetNbLatestEvents()
	n.addPathNode(spnm.id, pathNode, spnm.space)
	spnm.size++
	// New latest node
	if nbLatest == 0 {
		spnm.space.latestNodes = append(spnm.space.latestNodes, n)
	}
	return pathNode, true
}

func (spnm *SpacePathNodeMap) IsActive(pathNode m3path.PathNode) bool {
	n := spnm.space.GetNode(pathNode.P())
	if n != nil {
		return n.IsPathNodeActive(pathNode, spnm.space)
	}
	return false
}

/***************************************************************/
// Event Functions
/***************************************************************/

func (space *Space) CreateEvent(ctxType m3point.GrowthType, idx int, offset int, p m3point.Point, k EventColor) *Event {
	pnm := &SpacePathNodeMap{space, space.lastIdCounter, 0}
	space.lastIdCounter++
	ppd := m3point.GetPointPackData(space.env)
	ctx := m3path.MakePathContextDBFromGrowthContext(space.env, ppd.GetGrowthContextByTypeAndIndex(ctxType, idx), offset)
	e := Event{pnm.id, space, pnm,nil, space.currentTime, k, ctx}
	space.events[pnm.id] = &e
	ctx.InitRootNode(p)
	// TODO: Remove PathNodeMap need. Use DB
	pnm.AddPathNode(ctx.GetRootPathNode())
	e.node = space.GetNode(p)
	space.activeNodes.addNode(e.node)
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
