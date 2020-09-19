package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
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

