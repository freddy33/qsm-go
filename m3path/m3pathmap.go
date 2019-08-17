package m3path

import "github.com/freddy33/qsm-go/m3point"

type PathNodeMap interface {
	GetSize() int
	GetPathNode(p m3point.Point) (PathNode, bool)
	AddPathNode(pathNode PathNode) (PathNode, bool)
	IsActive(pathNode PathNode) bool
}

type SimplePathNodeMap map[m3point.Point]PathNode

const MaxHashConflicts = 6

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

func (pnm *SimplePathNodeMap) GetSize() int {
	return len(*pnm)
}

func (pnm *SimplePathNodeMap) GetPathNode(p m3point.Point) (PathNode, bool) {
	res, ok := (*pnm)[p]
	return res, ok
}

func (pnm *SimplePathNodeMap) AddPathNode(pathNode PathNode)  (PathNode, bool) {
	(*pnm)[pathNode.P()] = pathNode
	return pathNode, true
}

func (pnm *SimplePathNodeMap) IsActive(pathNode PathNode) bool {
	return pathNode.IsLatest()
}

/***************************************************************/
// PointHashPathNodeMap Functions
/***************************************************************/

func MakeHashPathNodeMap(initSize int) PathNodeMap {
	res := PointHashPathNodeMap{m3point.MakePointHashMap(initSize)}
	res.pointMap.SetMaxConflictsAllowed(8)
	return &res
}

func (hnm *PointHashPathNodeMap) GetSize() int {
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

