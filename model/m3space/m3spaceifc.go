package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type EventId int
type NodeEventId m3point.Int64Id

const (
	NilEvent = EventId(-1)
	MinMaxCoord = m3point.CInt(2 * 9)
)

type DistAndTime int

const ZeroDistAndTime = DistAndTime(0)

type EventColor uint8

const (
	RedEvent EventColor = 1 << iota
	GreenEvent
	BlueEvent
	YellowEvent
)

// TODO: This should be in the space data entry of the environment
var AllColors = [4]EventColor{RedEvent, GreenEvent, BlueEvent, YellowEvent}

type SpaceIfc interface {
	fmt.Stringer
	GetId() int
	GetName() string
	GetActiveThreshold() DistAndTime
	GetMaxTriosPerPoint() int
	GetMaxNodesPerPoint() int
	GetMaxTime() DistAndTime
	GetMaxCoord() m3point.CInt
	GetEvent(id EventId) EventIfc
	GetActiveEventsAt(atTime DistAndTime) []EventIfc
	GetSpaceTimeAt(atTime DistAndTime) SpaceTimeIfc
	CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int,
		creationTime DistAndTime, center m3point.Point, color EventColor) (EventIfc, error)
}

type SpaceTimeNodeVisitor interface {
	VisitNode(node SpaceTimeNodeIfc)
}

type SpaceTimeLinkVisitor interface {
	VisitLink(node SpaceTimeNodeIfc, srcPoint m3point.Point, connId m3point.ConnectionId)
}

type SpaceTimeIfc interface {
	fmt.Stringer
	GetSpace() SpaceIfc
	GetCurrentTime() DistAndTime
	GetActiveEvents() []EventIfc
	Next() SpaceTimeIfc
	GetNbActiveNodes() int
	GetNbActiveLinks() int
	VisitNodes(visitor SpaceTimeNodeVisitor)
	VisitLinks(visitor SpaceTimeLinkVisitor)
	GetDisplayState() string
}

type SpaceTimeNodeIfc interface {
	GetSpaceTime() SpaceTimeIfc
	GetPointId() m3path.PointId
	GetPoint() (*m3point.Point, error)

	IsEmpty() bool
	GetEventIds() []EventId

	HasRoot() bool
	GetLastAccessed() DistAndTime
	HowManyColors() uint8
	GetColorMask() uint8

	GetStateString() string
}

type EventIfc interface {
	fmt.Stringer
	GetId() EventId
	GetSpace() SpaceIfc
	GetPathContext() m3path.PathContext
	GetCreationTime() DistAndTime
	GetColor() EventColor
	GetCenterNode() NodeEventIfc
	GetActiveNodesAt(currentTime DistAndTime) ([]NodeEventIfc, error)
}

type NodeEventIfc interface {
	fmt.Stringer
	m3path.ConnectionStateIfc
	GetId() NodeEventId
	GetEventId() EventId

	GetPointId() m3path.PointId
	GetPoint() (*m3point.Point, error)

	GetPathNodeId() m3path.PathNodeId
	GetPathNode() (m3path.PathNode, error)

	GetCreationTime() DistAndTime
	GetD() DistAndTime
	GetColor() EventColor
}


