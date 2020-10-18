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

const (
	NewPathNodeId         m3point.Int64Id = -1
	InPoolId              m3point.Int64Id = -2
	LinkIdNotSet          m3point.Int64Id = -3
	DeadEndId             m3point.Int64Id = -4
	NextLinkIdNotAssigned m3point.Int64Id = -5
)

const (
	NilPointId PointId = -1
)

type PathContextId m3point.Int32Id
type PointId m3point.Int64Id
type PathNodeId m3point.Int64Id

type PathPoint struct {
	Id PointId
	P  m3point.Point
}

type PathContext interface {
	fmt.Stringer
	GetId() PathContextId
	GetGrowthCtx() m3point.GrowthContext
	GetGrowthType() m3point.GrowthType
	GetGrowthIndex() int
	GetGrowthOffset() int
	GetRootPathNode() PathNode

	GetMaxDist() int
	RequestNewMaxDist(requestDist int) error

	GetNumberOfNodesAt(dist int) int
	GetPathNodesAt(dist int) ([]PathNode, error)

	GetNumberOfNodesBetween(fromDist int, toDist int) int
	GetPathNodesBetween(fromDist, toDist int) ([]PathNode, error)

	DumpInfo() string
}

type ConnectionStateIfc interface {
	GetTrioIndex() m3point.TrioIndex
	GetTrioDetails(pointData m3point.PointPackDataIfc) *m3point.TrioDetails

	HasOpenConnections() bool
	IsFrom(connIdx int) bool
	IsNext(connIdx int) bool
	IsDeadEnd(connIdx int) bool
}

type PathNode interface {
	fmt.Stringer
	ConnectionStateIfc

	GetId() PathNodeId
	GetPathContext() PathContext

	IsRoot() bool
	P() m3point.Point
	D() int
}

var NilPathPoint = PathPoint{
	Id: NilPointId,
	P:  m3point.Origin,
}

func (pp PathPoint) String() string {
	return fmt.Sprintf("PP=%04d,%v", pp.Id, pp.P)
}

func (id PathContextId) MurmurHash() uint32 {
	return m3point.Int32Id(id).MurmurHash()
}

func (id PointId) MurmurHash() uint32 {
	return m3point.Int64Id(id).MurmurHash()
}

func (id PathNodeId) MurmurHash() uint32 {
	return m3point.Int64Id(id).MurmurHash()
}

func GetConnectionMaskValue(connectionMask uint16, connIdx int) uint16 {
	return (connectionMask >> uint16(connIdx*ConnectionMaskBits)) & SingleConnectionMask
}

func GetConnectionState(connectionMask uint16, connIdx int) ConnectionState {
	return ConnectionState(GetConnectionMaskValue(connectionMask, connIdx) & ConnectionStateMask)
}

func SetConnectionState(connectionMask uint16, connIdx int, state ConnectionState) uint16 {
	newConnMask := GetConnectionMaskValue(connectionMask, connIdx)
	// Zero what is not state mask bit
	newConnMask &^= ConnectionStateMask
	// Set the new state value
	newConnMask |= uint16(state)
	return newConnMask
}
