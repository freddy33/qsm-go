package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
)

type PathContext interface {
	fmt.Stringer
	GetGrowthCtx() m3point.GrowthContext
	GetGrowthOffset() int
	GetGrowthType() m3point.GrowthType
	GetGrowthIndex() int
	GetPathNodeMap() PathNodeMap
	CountAllPathNodes() int
	InitRootNode(center m3point.Point)
	GetRootPathNode() PathNode
	GetNumberOfOpenNodes() int
	GetAllOpenPathNodes() []PathNode
	MoveToNextNodes()
	PredictedNextOpenNodesLen() int
	dumpInfo() string
}

type PathNode interface {
	fmt.Stringer
	GetPathContext() PathContext

	IsRoot() bool
	IsLatest() bool
	P() m3point.Point
	D() int
	GetTrioIndex() m3point.TrioIndex

	HasOpenConnections() bool
	IsDeadEnd(connIdx int) bool
	SetDeadEnd(connIdx int)

	GetFrom() int64
	GetOtherFrom() int64
	GetNext(connIdx int) int64
	GetNextConnection(connId m3point.ConnectionId) int64
}


