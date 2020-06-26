package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/utils/m3db"
)

type GrowthContext interface {
	fmt.Stringer
	GetEnv() *m3db.QsmEnvironment
	GetId() int
	GetGrowthType() GrowthType
	GetGrowthIndex() int
	GetBaseDivByThree(mainPoint Point) uint64
	GetBaseTrioIndex(divByThree uint64, offset int) TrioIndex
}

type PathNodeBuilder interface {
	fmt.Stringer
	GetEnv() *m3db.QsmEnvironment
	GetCubeId() int
	GetTrioIndex() TrioIndex
	GetNextPathNodeBuilder(from Point, connId ConnectionId, offset int) (PathNodeBuilder, Point)
	dumpInfo() string
	verify()
}


