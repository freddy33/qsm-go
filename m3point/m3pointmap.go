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
	Clear()
	Range(f func(point Point, value interface{}) bool, nbProc int)
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
	nbElements          []int
	maxConflicts        int
	mutexes             []*sync.Mutex
	data                []*pointHashMapEntry
}

func MakePointHashMap(mapSize int, segments int) PointMap {
	res := new(pointHashMap)
	res.fullLocked = false
	res.maxConflictsAllowed = DefaultMaxHashConflicts
	res.showedError = false
	res.nbElements = make([]int, segments)
	res.data = make([]*pointHashMapEntry, mapSize)
	res.mutexes = make([]*sync.Mutex, segments)
	for i := 0; i < segments; i++ {
		res.mutexes[i] = new(sync.Mutex)
	}
	return res
}

func (phm *pointHashMap) Size() int {
	res := 0
	for _, n := range phm.nbElements {
		res += n
	}
	return res
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

func (phm *pointHashMap) internalPut(p *Point, val interface{}, overrideValue bool) (interface{}, bool) {
	if p == nil {
		return nil, false
	}

	rp := *p
	key := rp.Hash(len(phm.data))

	segmentIdx := key % len(phm.mutexes)
	mutex := phm.mutexes[segmentIdx]
	locked := false

	// TODO: Moved to lock free insert using atomic.CompareAndSwapPointer
	if phm.fullLocked {
		mutex.Lock()
		defer mutex.Unlock()
		locked = true
	}

	entry := phm.data[key]
	if entry == nil {
		if !locked {
			mutex.Lock()
			defer mutex.Unlock()
			locked = true
		}
		entry = phm.data[key]
		if entry == nil {
			entry = new(pointHashMapEntry)
			entry.point = rp
			entry.value = val
			entry.next = nil
			phm.data[key] = entry
			phm.nbElements[segmentIdx]++
			if overrideValue {
				return nil, true
			} else {
				return entry.value, true
			}
		}
	}
	deepness := 0
	for {
		if entry.point == rp {
			if overrideValue {
				oldVal := entry.value
				entry.value = val
				return oldVal, false
			} else {
				return entry.value, false
			}
		}
		if entry.next == nil {
			if !locked {
				mutex.Lock()
				defer mutex.Unlock()
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
				phm.nbElements[segmentIdx]++
				if overrideValue {
					return nil, true
				} else {
					return newEntry.value, true
				}
			}
		}
		deepness++
		if !phm.showedError && deepness > phm.maxConflictsAllowed {
			phm.showedError = true
			Log.Errorf("The size %d of map is too small to contain %d objects since max conflicts allowed set to %d and got here %d", len(phm.data), phm.Size(), phm.maxConflictsAllowed, deepness)
		}
		entry = entry.next
	}
}

func (phm *pointHashMap) Clear() {
	for i := 0; i < len(phm.mutexes); i++ {
		mutex := phm.mutexes[i]
		mutex.Lock()
		defer mutex.Unlock()
	}
	// Just nilify all link list
	for i := 0; i < len(phm.data); i++ {
		entry := phm.data[i]
		for {
			if entry == nil {
				break
			}
			toClear := entry
			entry = entry.next
			toClear.next = nil
			toClear.value = nil
		}
		phm.data[i] = nil
	}
	for i := 0; i < len(phm.mutexes); i++ {
		phm.nbElements[i] = 0
	}
}

func (phm *pointHashMap) Range(f func(point Point, value interface{}) bool, nbProc int) {
	dataSize := len(phm.data)
	segSize := dataSize / nbProc
	if segSize < 2 {
		// Simple pass
		for i := 0; i < dataSize; i++ {
			entry := phm.data[i]
			for {
				if entry == nil {
					break
				}
				if f(entry.point, entry.value) {
					return
				}
				entry = entry.next
			}
		}
	} else {
		// Parallelize
		wg := new(sync.WaitGroup)
		wg.Add(nbProc)
		for segId := 0; segId < nbProc; segId++ {
			startIdx := dataSize * segId
			endIdx := startIdx + dataSize
			if endIdx > dataSize {
				endIdx = dataSize
			}
			go func() {
				for i := startIdx; i < endIdx; i++ {
					entry := phm.data[i]
					for {
						if entry == nil {
							break
						}
						if f(entry.point, entry.value) {
							// TODO: find a way to stop all go routines
							return
						}
						entry = entry.next
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
