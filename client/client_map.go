package client

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"unsafe"
)

type ClientPathNodeMap interface {
	Size() int
	GetPathNodeById(id m3point.Int64Id) *PathNodeCl
	GetPathNode(p m3point.Point) *PathNodeCl
	AddPathNode(pathNode *PathNodeCl) (*PathNodeCl, bool)
	Clear()
	RangePerId(visit func(id m3point.Int64Id, pn *PathNodeCl) bool, rc *m3point.RangeContext)
	RangePerPoint(visit func(point m3point.Point, pn *PathNodeCl) bool, rc *m3point.RangeContext)
}

type ClientPointHashPathNodeMap struct {
	idMap    *m3point.NonBlockConcurrentMap
	pointMap m3point.PointMap
}

/***************************************************************/
// ServerPointHashPathNodeMap Functions
/***************************************************************/

// TODO: This should be in path data entry of the env
var nbParallelProcesses = 8

func MakeHashPathNodeMap(initSize int) ClientPathNodeMap {
	res := ClientPointHashPathNodeMap{
		idMap:    m3point.MakeNonBlockConcurrentMap(initSize),
		pointMap: m3point.MakePointHashMap(initSize),
	}
	return &res
}

func (hnm *ClientPointHashPathNodeMap) Size() int {
	return hnm.pointMap.Size()
}

func (hnm *ClientPointHashPathNodeMap) GetPathNodeById(id m3point.Int64Id) *PathNodeCl {
	pn, ok := hnm.idMap.Load(id)
	if ok {
		return (*PathNodeCl)(pn)
	}
	return nil
}

func (hnm *ClientPointHashPathNodeMap) GetPathNode(p m3point.Point) *PathNodeCl {
	pn, ok := hnm.pointMap.Get(p)
	if ok {
		return (*PathNodeCl)(pn)
	}
	return nil
}

func (hnm *ClientPointHashPathNodeMap) AddPathNode(pathNode *PathNodeCl) (*PathNodeCl, bool) {
	if pathNode.id < 0 {
		Log.Fatalf("cannot add unsaved node %s", pathNode.String())
		return nil, false
	}
	p := pathNode.P()
	up := unsafe.Pointer(pathNode)
	pn, inserted := hnm.pointMap.LoadOrStore(p, up)
	if inserted {
		hnm.idMap.Store(pathNode.id, up)
	}
	return (*PathNodeCl)(pn), inserted
}

func (hnm *ClientPointHashPathNodeMap) Clear() {
	rc := m3point.MakeRangeContext(false, nbParallelProcesses, Log)
	hnm.pointMap.Range(func(point m3point.Point, value unsafe.Pointer) bool {
		pn := (*PathNodeCl)(value)
		pn.release()
		return false
	}, rc)
	hnm.pointMap = m3point.MakePointHashMap(hnm.pointMap.InitSize())
}

func (hnm *ClientPointHashPathNodeMap) RangePerId(visit func(id m3point.Int64Id, pn *PathNodeCl) bool, rc *m3point.RangeContext) {
	hnm.idMap.Range(func(key m3point.MurmurKey, value unsafe.Pointer) bool {
		return visit(key.(m3point.Int64Id), (*PathNodeCl)(value))
	}, rc)
}

func (hnm *ClientPointHashPathNodeMap) RangePerPoint(visit func(point m3point.Point, pn *PathNodeCl) bool, rc *m3point.RangeContext) {
	hnm.pointMap.Range(func(point m3point.Point, value unsafe.Pointer) bool {
		return visit(point, (*PathNodeCl)(value))
	}, rc)
}
