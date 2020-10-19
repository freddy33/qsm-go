package m3path

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"unsafe"
)

type PathPointMap struct {
	idMap    *m3point.NonBlockConcurrentMap
	pointMap *m3point.NonBlockConcurrentMap
}

type PathPackDataIfc interface {
	m3util.QsmDataPack
	GetPathCtx(id PathContextId) PathContext
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

func MakePathPointMap(initSize int) *PathPointMap {
	res := PathPointMap{
		idMap:    m3point.MakeNonBlockConcurrentMap(initSize),
		pointMap: m3point.MakeNonBlockConcurrentMap(initSize),
	}
	return &res
}

func (ppm *PathPointMap) GetById(pointId PointId) (*PathPoint, bool) {
	pp, ok := ppm.idMap.Load(pointId)
	if ok {
		return (*PathPoint)(pp), true
	}
	return (*PathPoint)(nil), false
}

func (ppm *PathPointMap) GetByPoint(point m3point.Point) (*PathPoint, bool) {
	pp, ok := ppm.pointMap.Load(point)
	if ok {
		return (*PathPoint)(pp), true
	}
	return (*PathPoint)(nil), false
}

func (ppm *PathPointMap) AddToMap(pointId PointId, point m3point.Point) *PathPoint {
	newPathPoint := &PathPoint{
		Id: pointId,
		P:  point,
	}
	actualPP, _ := ppm.idMap.LoadOrStore(pointId, unsafe.Pointer(newPathPoint))
	actualPathPoint := (*PathPoint)(actualPP)
	ppm.pointMap.Store((*actualPathPoint).P, actualPP)
	return actualPathPoint
}


