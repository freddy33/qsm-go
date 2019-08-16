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
	size int
	data []*[]PathNode
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
	res := PointHashPathNodeMap{ 0,make([]*[]PathNode, initSize)}
	return &res
}

func (hnm *PointHashPathNodeMap) GetSize() int {
	return hnm.size
}

func (hnm *PointHashPathNodeMap) GetPathNode(p m3point.Point) (PathNode, bool) {
	key := p.Hash(len(hnm.data))
	l := hnm.data[key]
	for _, pn := range *l {
		if pn != nil && pn.P() == p {
			return pn, true
		}
	}
	return nil, false
}

func (hnm *PointHashPathNodeMap) AddPathNode(pathNode PathNode) (PathNode, bool) {
	p := pathNode.P()
	key := p.Hash(len(hnm.data))
	l := hnm.data[key]
	if l == nil {
		newL := make([]PathNode, 3)
		hnm.data[key] = &newL
		// retrieving from array to limit race issue
		l = hnm.data[key]
	}
	for i := 0; i < MaxHashConflicts; i++ {
		if i >= len(*l) {
			*(hnm.data[key]) = append(*(hnm.data[key]), pathNode)
			hnm.size++
			return pathNode, true
		}
		pn := (*l)[i]
		if pn == nil {
			(*l)[i] = pathNode
			hnm.size++
			return pathNode, true
		} else if pn.P() == p {
			return pn, false
		}
	}
	Log.Errorf("Too many conflicts with cap %d and current size %d", len(hnm.data), hnm.size)
	return nil, false
}

func (*PointHashPathNodeMap) IsActive(pathNode PathNode) bool {
	return pathNode.IsLatest()
}

