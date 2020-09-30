package m3path

import "github.com/freddy33/qsm-go/model/m3point"

type PathNodeMap interface {
	Size() int
	GetPathNode(p m3point.Point) PathNode
	AddPathNode(pathNode PathNode) (PathNode, bool)
	Clear()
	Range(f func(point m3point.Point, pn PathNode) bool, nbProc int)
}

type SimplePathNodeMap map[m3point.Point]PathNode

type PointHashPathNodeMap struct {
	pointMap m3point.PointMap
}

/***************************************************************/
// SimplePathNodeMap Functions
/***************************************************************/

func MakeSimplePathNodeMap(initSize int) PathNodeMap {
	res := SimplePathNodeMap(make(map[m3point.Point]PathNode, initSize))
	return &res
}

func (pnm *SimplePathNodeMap) Size() int {
	return len(*pnm)
}

func (pnm *SimplePathNodeMap) GetPathNode(p m3point.Point) PathNode {
	res, ok := (*pnm)[p]
	if !ok {
		return nil
	}
	return res
}

func (pnm *SimplePathNodeMap) AddPathNode(pathNode PathNode)  (PathNode, bool) {
	p := pathNode.P()
	res, ok := (*pnm)[p]
	if !ok {
		res = pathNode
		(*pnm)[p] = pathNode
	}
	return res, !ok
}

func (pnm *SimplePathNodeMap) Clear() {
	for k, _ := range *pnm {
		delete(*pnm, k)
	}
}

func (pnm *SimplePathNodeMap) Range(f func(point m3point.Point, pn PathNode) bool, nbProc int) {
	for k, v := range *pnm {
		if f(k, v) {
			return
		}
	}
}

/***************************************************************/
// PointHashPathNodeMap Functions
/***************************************************************/

func MakeHashPathNodeMap(initSize int) PathNodeMap {
	res := PointHashPathNodeMap{m3point.MakePointHashMap(initSize, 16)}
	res.pointMap.SetMaxConflictsAllowed(8)
	return &res
}

func (hnm *PointHashPathNodeMap) Size() int {
	return hnm.pointMap.Size()
}

func (hnm *PointHashPathNodeMap) GetPathNode(p m3point.Point) PathNode {
	pn, ok := hnm.pointMap.Get(&p)
	if ok {
		return pn.(PathNode)
	}
	return nil
}

func (hnm *PointHashPathNodeMap) AddPathNode(pathNode PathNode) (PathNode, bool) {
	p := pathNode.P()
	pn, inserted := hnm.pointMap.LoadOrStore(&p, pathNode)
	return pn.(PathNode), inserted
}

func (hnm *PointHashPathNodeMap) Clear() {
	hnm.pointMap.Clear()
}

func (hnm *PointHashPathNodeMap) Range(f func(point m3point.Point, pn PathNode) bool, nbProc int) {
	hnm.pointMap.Range(func(point m3point.Point, value interface{}) bool {
		return f(point, value.(PathNode))
	}, nbProc)
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

