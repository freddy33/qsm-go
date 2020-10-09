package pathdb

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"unsafe"
)

type ServerPathNodeMap interface {
	Size() int
	GetPathNode(p m3point.Point) *PathNodeDb
	AddPathNode(pathNode *PathNodeDb) (*PathNodeDb, bool)
	Clear()
	Range(visit func(point m3point.Point, pn *PathNodeDb) bool, rc *m3point.RangeContext)
}

type SimplePathNodeMap map[m3point.Point]*PathNodeDb

type ServerPointHashPathNodeMap struct {
	pointMap m3point.PointMap
}

/***************************************************************/
// SimplePathNodeMap Functions
/***************************************************************/

func MakeSimplePathNodeMap(initSize int) ServerPathNodeMap {
	res := SimplePathNodeMap(make(map[m3point.Point]*PathNodeDb, initSize))
	return &res
}

func (pnm *SimplePathNodeMap) Size() int {
	return len(*pnm)
}

func (pnm *SimplePathNodeMap) GetPathNode(p m3point.Point) *PathNodeDb {
	res, ok := (*pnm)[p]
	if !ok {
		return nil
	}
	return res
}

func (pnm *SimplePathNodeMap) AddPathNode(pathNode *PathNodeDb) (*PathNodeDb, bool) {
	p := pathNode.P()
	res, ok := (*pnm)[p]
	if !ok {
		res = pathNode
		(*pnm)[p] = pathNode
	}
	return res, !ok
}

func (pnm *SimplePathNodeMap) Clear() {
	for k, pn := range *pnm {
		pn.release()
		delete(*pnm, k)
	}
}

func (pnm *SimplePathNodeMap) Range(visit func(point m3point.Point, pn *PathNodeDb) bool, rc *m3point.RangeContext) {
	rc.Wg.Add(1)
	go func() {
		defer rc.Wg.Done()
		for k, v := range *pnm {
			if visit(k, v) {
				return
			}
		}
	}()
	if !rc.IsAsync() {
		rc.Wait()
	}
}

/***************************************************************/
// ServerPointHashPathNodeMap Functions
/***************************************************************/

func MakeHashPathNodeMap(initSize int) ServerPathNodeMap {
	res := ServerPointHashPathNodeMap{m3point.MakePointHashMap(initSize)}
	return &res
}

func (hnm *ServerPointHashPathNodeMap) Size() int {
	return hnm.pointMap.Size()
}

func (hnm *ServerPointHashPathNodeMap) GetPathNode(p m3point.Point) *PathNodeDb {
	pn, ok := hnm.pointMap.Get(p)
	if ok {
		return (*PathNodeDb)(pn)
	}
	return nil
}

func (hnm *ServerPointHashPathNodeMap) AddPathNode(pathNode *PathNodeDb) (*PathNodeDb, bool) {
	p := pathNode.P()
	pn, inserted := hnm.pointMap.LoadOrStore(p, unsafe.Pointer(pathNode))
	return (*PathNodeDb)(pn), inserted
}

func (hnm *ServerPointHashPathNodeMap) Clear() {
	rc := m3point.MakeRangeContext(false, nbParallelProcesses, Log)
	hnm.pointMap.Range(func(point m3point.Point, value unsafe.Pointer) bool {
		pn := (*PathNodeDb)(value)
		pn.release()
		return false
	}, rc)
	hnm.pointMap = m3point.MakePointHashMap(hnm.pointMap.InitSize())
}

func (hnm *ServerPointHashPathNodeMap) Range(visit func(point m3point.Point, pn *PathNodeDb) bool, rc *m3point.RangeContext) {
	hnm.pointMap.Range(func(point m3point.Point, value unsafe.Pointer) bool {
		return visit(point, (*PathNodeDb)(value))
	}, rc)
}

