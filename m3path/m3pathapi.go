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
	InitRootNode(center m3point.Point)
	GetRootPathNode() PathNode
	GetNumberOfOpenNodes() int
	GetAllOpenPathNodes() []PathNode
	MoveToNextNodes()
	PredictedNextOpenNodesLen() int
	dumpInfo() string
}

type PathLink interface {
	fmt.Stringer
	GetSrc() PathNode
	GetConnId() m3point.ConnectionId
	HasDestination() bool
	IsDeadEnd() bool
	SetDeadEnd()
	createDstNode(pathBuilder m3point.PathNodeBuilder) (PathNode, bool, m3point.PathNodeBuilder)
	dumpInfo(ident int) string
}

type PathNode interface {
	fmt.Stringer
	GetPathContext() *BasePathContext
	IsEnd() bool
	IsRoot() bool
	IsLatest() bool
	P() m3point.Point
	D() int
	GetTrioIndex() m3point.TrioIndex
	GetFrom() PathLink
	GetOtherFrom() PathLink
	GetNext(i int) PathLink
	GetNextConnection(connId m3point.ConnectionId) PathLink

	calcDist() int
	addPathLink(connId m3point.ConnectionId) (PathLink, bool)
	setOtherFrom(pl PathLink)
	dumpInfo(ident int) string
}


