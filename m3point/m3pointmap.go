package m3point

import "sync"

const DefaultMaxHashConflicts = 8

type PointMap interface {
	Size() int
	Get(p *Point) (interface{}, bool)
	Put(p *Point, val interface{}) (interface{}, bool)
	LoadOrStore(p *Point, val interface{}) (interface{}, bool)
	GetMaxConflictsAllowed() int
	GetCurrentMaxConflicts() int
	SetMaxConflictsAllowed(max int)
}

type pointHashMapEntry struct {
	point Point
	value interface{}
	next  *pointHashMapEntry
}

type pointHashMap struct {
	fullLocked          bool
	maxConflictsAllowed int
	showedError         bool
	nbElements          int
	maxConflicts        int
	mutex               sync.Mutex
	conflictsMutex      sync.Mutex
	data                []*pointHashMapEntry
}

func MakePointHashMap(mapSize int) PointMap {
	res := new(pointHashMap)
	res.fullLocked = false
	res.maxConflictsAllowed = DefaultMaxHashConflicts
	res.showedError = false
	res.nbElements = 0
	res.data = make([]*pointHashMapEntry, mapSize)
	return res
}

func (phm *pointHashMap) Size() int {
	return phm.nbElements
}

func (phm *pointHashMap) GetMaxConflictsAllowed() int {
	return phm.maxConflictsAllowed
}

func (phm *pointHashMap) SetMaxConflictsAllowed(max int) {
	phm.maxConflictsAllowed = max
}

func (phm *pointHashMap) GetCurrentMaxConflicts() int {
	return phm.maxConflicts
}

func (phm *pointHashMap) Get(p *Point) (interface{}, bool) {
	if p == nil {
		return nil, false
	}
	key := p.Hash(len(phm.data))
	entry := phm.data[key]
	rp := *p
	for {
		if entry == nil {
			return nil, false
		}
		if entry.point == rp {
			return entry.value, true
		}
		entry = entry.next
	}
}

func (phm *pointHashMap) Put(p *Point, val interface{}) (interface{}, bool) {
	return phm.internalPut(p, val, true)
}

func (phm *pointHashMap) LoadOrStore(p *Point, val interface{}) (interface{}, bool) {
	return phm.internalPut(p, val, false)
}

func (phm *pointHashMap) internalPut(p *Point, val interface{}, insert bool) (interface{}, bool) {
	if p == nil {
		return nil, false
	}

	locked := false

	if phm.fullLocked {
		phm.mutex.Lock()
		defer phm.mutex.Unlock()
		locked = true
	}

	rp := *p
	key := rp.Hash(len(phm.data))
	entry := phm.data[key]
	if entry == nil {
		if !locked {
			phm.mutex.Lock()
			defer phm.mutex.Unlock()
			locked = true
		}
		entry = phm.data[key]
		if entry == nil {
			entry = new(pointHashMapEntry)
			entry.point = rp
			entry.value = val
			entry.next = nil
			phm.data[key] = entry
			phm.nbElements++
			if insert {
				return nil, true
			} else {
				return entry.value, true
			}
		}
	}
	deepness := 0
	for {
		if entry.point == rp {
			if insert {
				oldVal := entry.value
				entry.value = val
				return oldVal, false
			} else {
				return entry.value, false
			}
		}
		if entry.next == nil {
			if !locked {
				phm.mutex.Lock()
				defer phm.mutex.Unlock()
				locked = true
			}
			if entry.next == nil {
				deepness++
				if deepness > phm.maxConflicts {
					phm.maxConflicts = deepness
				}
				newEntry := new(pointHashMapEntry)
				newEntry.point = rp
				newEntry.value = val
				newEntry.next = nil
				entry.next = newEntry
				phm.nbElements++
				if insert {
					return nil, true
				} else {
					return newEntry.value, true
				}
			}
		}
		deepness++
		if !phm.showedError && deepness > phm.maxConflictsAllowed {
			phm.conflictsMutex.Lock()
			defer phm.conflictsMutex.Unlock()
			if !phm.showedError {
				Log.Errorf("The size %d of map is too small to contain %d objects since max conflicts allowed set to %d and got here %d", len(phm.data), phm.nbElements, phm.maxConflictsAllowed, deepness)
				phm.showedError = true
			}
		}
		entry = entry.next
	}
}
