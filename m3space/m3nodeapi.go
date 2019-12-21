package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
)

type Node interface {
	fmt.Stringer
	GetPoint() *m3point.Point

	IsEmpty() bool
	HasRoot() bool

	GetNbEvents() int
	GetNbLatestEvents() int
	GetLatestEventIds() []EventID
	GetNbActiveEvents(space *Space) int
	GetActiveEventIds(space *Space) []EventID
	GetActiveLinks(space *Space) NodeLinkList

	IsEventAlreadyPresent(id EventID) bool

	GetPathNode(id EventID) m3path.PathNode

	GetAccessed(evt *Event) DistAndTime

	GetLastAccessed(space *Space) DistAndTime

	GetEventDistFromCurrent(evt *Event) DistAndTime
	GetEventForPathNode(pathNode m3path.PathNode, space *Space) *Event
	IsPathNodeActive(pathNode m3path.PathNode, space *Space) bool

	HowManyColors(space *Space) uint8
	GetColorMask(space *Space) uint8

	IsActive(space *Space) bool
	IsOld(space *Space) bool
	IsDead(space *Space) bool

	GetStateString(space *Space) string

	addPathNode(id EventID, pn m3path.PathNode, space *Space)
}

type NodeEvent interface {
	GetEventId() EventID
	GetPathNodeId() int64
	GetPathNode() m3path.PathNode

	GetAccessedTime() DistAndTime
	GetDistFromCurrent(space *Space) DistAndTime

	IsLatest() bool
	IsRoot(evt *Event) bool
	IsActive(space *Space) bool
	IsActiveNext(space *Space) bool
	IsOld(space *Space) bool
	IsDead(space *Space) bool
}

type NodeLink interface {
	GetConnId() m3point.ConnectionId
	GetSrc() m3point.Point
}

type NodeList []Node
type NodeLinkList []NodeLink
