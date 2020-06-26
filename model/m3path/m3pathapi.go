package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
)

type PathContext interface {
	fmt.Stringer
	GetId() int
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
	GetId() int64
	GetPathContext() PathContext

	IsRoot() bool
	IsLatest() bool
	P() m3point.Point
	D() int
	GetTrioIndex() m3point.TrioIndex

	HasOpenConnections() bool
	IsFrom(connIdx int) bool
	IsNext(connIdx int) bool
	IsDeadEnd(connIdx int) bool

	GetTrioDetails() *m3point.TrioDetails
}


