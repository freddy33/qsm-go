package m3space

import (
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type EventId int

const (
	NilEvent = EventId(-1)
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

type EventNodeIfc interface {
	GetId() int64
	GetEventId() EventId
	GetPointId() int64
	GetCreationTime() DistAndTime
	GetD() DistAndTime
	GetPoint() (*m3point.Point, error)
	GetPathNode() (m3path.PathNode, error)
}

type EventIfc interface {
	GetId() EventId
	GetSpace() SpaceIfc
	GetPathContext() m3path.PathContext
	GetCreationTime() DistAndTime
	GetColor() EventColor
	GetCenterNode() EventNodeIfc
}

type Event struct {
	Id          EventId
	space       *Space
	pathNodeMap m3path.PathNodeMap
	node        Node
	created     DistAndTime
	color       EventColor
	PathContext m3path.PathContext
}

func (evt *Event) GetId() EventId {
	return evt.Id
}

func (evt *Event) GetSpace() SpaceIfc {
	return evt.space
}

func (evt *Event) GetPathContext() m3path.PathContext {
	return evt.PathContext
}

func (evt *Event) GetCreationTime() DistAndTime {
	return evt.created
}

func (evt *Event) GetColor() EventColor {
	return evt.color
}

func (evt *Event) GetCenterNode() EventNodeIfc {
	return evt.node.(*BaseNode).head.cur
}

type SpacePathNodeMap struct {
	space *Space
	id    EventId
	size  int
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

func (spnm *SpacePathNodeMap) Size() int {
	return spnm.size
}

func (spnm *SpacePathNodeMap) GetPathNode(p m3point.Point) m3path.PathNode {
	res, ok := spnm.space.nodesMap.Load(p)
	if ok {
		pathNode, err := res.(Node).GetPathNode(spnm.id)
		if err != nil {
			Log.Error(err)
		}
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

func (space *Space) CreateEventAtZero(ctxType m3point.GrowthType, idx int, offset int, p m3point.Point, k EventColor) *Event {
	pnm := &SpacePathNodeMap{space, space.lastIdCounter, 0}
	space.lastIdCounter++
	ppd := space.GetPointPackData()
	ctx := space.GetPathPackData().CreatePathCtxFromAttributes(ppd.GetGrowthContextByTypeAndIndex(ctxType, idx), offset, p)
	e := Event{pnm.id, space, pnm,nil, space.CurrentTime, k, ctx}
	space.events[pnm.id] = &e
	//ctx.InitRootNode(p)
	// TODO: Remove PathNodeMap need. Use DB
	pnm.AddPathNode(ctx.GetRootPathNode())
	e.node = space.GetNode(p)
	space.ActiveNodes.addNode(e.node)
	return &e
}

func (space *Space) CreateEventFromColor(p m3point.Point, k EventColor) *Event {
	idx, offset := getIndexAndOffsetForColor(k)
	return space.CreateEventAtZero(8, idx, offset, p, k)
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
	return evt.space.CurrentTime - evt.created
}
