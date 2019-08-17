package m3path

import "github.com/freddy33/qsm-go/m3point"

type PathNodeMap interface {
	Size() int
	GetPathNode(p m3point.Point) (PathNode, bool)
	AddPathNode(pathNode PathNode) (PathNode, bool)
	IsActive(pathNode PathNode) bool
	Clear()
	Range(f func(point m3point.Point, pn PathNode) bool)
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

func (pnm *SimplePathNodeMap) GetPathNode(p m3point.Point) (PathNode, bool) {
	res, ok := (*pnm)[p]
	return res, ok
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

func (pnm *SimplePathNodeMap) IsActive(pathNode PathNode) bool {
	return pathNode.IsLatest()
}

func (pnm *SimplePathNodeMap) Clear() {
	for k, _ := range *pnm {
		delete(*pnm, k)
	}
}

func (pnm *SimplePathNodeMap) Range(f func(point m3point.Point, pn PathNode) bool) {
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

func (hnm *PointHashPathNodeMap) GetPathNode(p m3point.Point) (PathNode, bool) {
	pn, ok := hnm.pointMap.Get(&p)
	if ok {
		return pn.(PathNode), true
	}
	return nil, false
}

func (hnm *PointHashPathNodeMap) AddPathNode(pathNode PathNode) (PathNode, bool) {
	p := pathNode.P()
	pn, inserted := hnm.pointMap.LoadOrStore(&p, pathNode)
	return pn.(PathNode), inserted
}

func (*PointHashPathNodeMap) IsActive(pathNode PathNode) bool {
	return pathNode.IsLatest()
}

func (hnm *PointHashPathNodeMap) Clear() {
	hnm.pointMap.Clear()
}

func (hnm *PointHashPathNodeMap) Range(f func(point m3point.Point, pn PathNode) bool) {
	hnm.pointMap.Range(func(point m3point.Point, value interface{}) bool {
		return f(point, value.(PathNode))
	})
}
