package m3point

import "fmt"

type GrowthContext interface {
	fmt.Stringer
	GetId() int
	GetGrowthType() GrowthType
	GetGrowthIndex() int
	GetBaseDivByThree(mainPoint Point) uint64
	GetBaseTrioDetails(mainPoint Point, offset int) *TrioDetails
	GetBaseTrioIndex(divByThree uint64, offset int) TrioIndex
}

type PathNodeBuilder interface {
	fmt.Stringer
	GetCubeId() int
	GetTrioIndex() TrioIndex
	GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point)
	dumpInfo() string
	verify()
}


