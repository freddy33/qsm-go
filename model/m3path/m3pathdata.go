package m3path

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type PathPackDataIfc interface {
	m3util.QsmDataPack
	GetPathCtx(id int) PathContext
	GetPathCtxFromAttributes(growthType m3point.GrowthType, growthIndex int, growthOffset int) (PathContext, error)
}

func CalculatePredictedSize(growthType m3point.GrowthType, d int) int {
	if d == 0 {
		return 3
	}
	if d == 1 {
		return 6
	}

	buffer := float32(1.02)
	df := float32(d)
	if growthType == m3point.GrowthType(8) {
		return int((1.775*df*df - 2.497*df + 5.039) * buffer)
	} else if growthType == m3point.GrowthType(2) {
		return int((1.445*df*df - 0.065*df - 0.377) * buffer)
	}
	// TODO: Find trend lines for other context types
	return int((1.775*df*df - 2.497*df + 5.039) * buffer)
}

