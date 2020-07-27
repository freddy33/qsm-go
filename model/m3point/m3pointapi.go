package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/utils/m3util"
)

type GrowthContext interface {
	fmt.Stringer
	GetEnv() m3util.QsmEnvironment
	GetId() int
	GetGrowthType() GrowthType
	GetGrowthIndex() int
	GetBaseDivByThree(mainPoint Point) uint64
	GetBaseTrioIndex(ppd PointPackDataIfc, divByThree uint64, offset int) TrioIndex
}

type PathNodeBuilder interface {
	fmt.Stringer
	GetEnv() m3util.QsmEnvironment
	GetCubeId() int
	GetTrioIndex() TrioIndex
	GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point)
	DumpInfo() string
	Verify()
}


