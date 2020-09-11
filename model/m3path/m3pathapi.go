package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

var Log = m3util.NewLogger("m3path", m3util.INFO)

const NbConnections = 3

type ConnectionState uint16

const (
	ConnectionMaskBits   = 4
	ConnectionStateMask  = uint16(0x0003)
	SingleConnectionMask = uint16(0x000F)
)
const (
	ConnectionNotSet  ConnectionState = 0x0000
	ConnectionFrom    ConnectionState = 0x0001
	ConnectionNext    ConnectionState = 0x0002
	ConnectionBlocked ConnectionState = 0x0003
	// Extra states possible as mask
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
	DumpInfo() string
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

func GetConnectionMaskValue(connectionMask uint16, connIdx int) uint16 {
	return (connectionMask >> uint16(connIdx*ConnectionMaskBits)) & SingleConnectionMask
}

func GetConnectionState(connectionMask uint16, connIdx int) ConnectionState {
	return ConnectionState(GetConnectionMaskValue(connectionMask, connIdx) & ConnectionStateMask)
}
