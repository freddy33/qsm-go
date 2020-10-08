package m3point

import "unsafe"

const DefaultMaxHashConflicts = 8



type PointMap interface {
	InitSize() int
	Size() int
	Get(p Point) (value unsafe.Pointer, loaded bool)
	Put(p Point, val unsafe.Pointer) unsafe.Pointer
	LoadOrStore(p Point, val unsafe.Pointer) (actual unsafe.Pointer, inserted bool)
	Range(visit func(point Point, value unsafe.Pointer) bool, rc *RangeContext)
}

type PointHashMap struct {
	maxConflictsAllowed int
	concMap *NonBlockConcurrentMap
}

func MakePointHashMap(initSize int) PointMap {
	res := new(PointHashMap)
	res.concMap = MakeNonBlockConcurrentMap(initSize)
	res.maxConflictsAllowed = DefaultMaxHashConflicts
	return res
}

func (phm *PointHashMap) InitSize() int {
	return phm.concMap.InitSize()
}

func (phm *PointHashMap) Size() int {
	return phm.concMap.Size()
}

func (phm *PointHashMap) GetMaxConflictsAllowed() int {
	return phm.maxConflictsAllowed
}

func (phm *PointHashMap) SetMaxConflictsAllowed(max int) {
	phm.maxConflictsAllowed = max
}

func (phm *PointHashMap) GetCurrentMaxConflicts() int {
	return phm.concMap.mHashConflicts
}

func (phm *PointHashMap) Get(p Point) (unsafe.Pointer, bool) {
	return phm.concMap.Load(p)
}

func (phm *PointHashMap) Put(p Point, val unsafe.Pointer) unsafe.Pointer {
	return phm.concMap.Store(p, val)
}

func (phm *PointHashMap) LoadOrStore(p Point, val unsafe.Pointer) (actual unsafe.Pointer, inserted bool) {
	return phm.concMap.LoadOrStore(p, val)
}

func (phm *PointHashMap) Range(visit func(point Point, value unsafe.Pointer) bool, rc *RangeContext) {
	phm.concMap.Range(func(key MurmurKey, value unsafe.Pointer) bool {
		return visit(key.(Point), value)
	}, rc)
}
