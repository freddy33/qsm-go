package m3point

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestPointMapBasic(t *testing.T) {
	m := MakePointHashMap(10)
	p1 := Point{1, 2, 3}
	val23 := 23
	nilPointer := unsafe.Pointer((*int)(nil))
	uPval23 := unsafe.Pointer(&val23)
	o1 := m.Put(p1, uPval23)
	good := assert.Equal(t, nilPointer, o1) &&
		assert.Equal(t, 1, m.Size())
	if !good {
		return
	}
	p2 := Point{1, 2, 3}
	o2, b2 := m.Get(p2)
	good = assert.Equal(t, 23, *((*int)(o2))) &&
		assert.Equal(t, true, b2)
	if !good {
		return
	}
	p2[0] = 2
	o3, b3 := m.Get(p2)
	good = assert.Equal(t, nilPointer, o3) &&
		assert.Equal(t, false, b3) &&
		assert.Equal(t, 1, m.Size())
	if !good {
		return
	}
	val24 := 24
	uPVal24 := unsafe.Pointer(&val24)
	o4 := m.Put(p2, uPVal24)
	good = assert.Equal(t, nilPointer, o4) &&
		assert.Equal(t, 2, m.Size())
	if !good {
		return
	}

	val25 := 25
	oldVal23 := m.Put(p1, unsafe.Pointer(&val25))
	good = assert.Equal(t, 23, *((*int)(oldVal23))) &&
		assert.Equal(t, uPval23, oldVal23) &&
		assert.Equal(t, 2, m.Size())
	if !good {
		return
	}

	o25, b25 := m.Get(p1)
	good = assert.Equal(t, 25, *((*int)(o25))) &&
		assert.Equal(t, true, b25)
	if !good {
		return
	}

	p3 := Point{3,3,3}
	val26 := 26
	uPVal26 := unsafe.Pointer(&val26)
	actualVal26, inserted := m.LoadOrStore(p3, uPVal26)
	good = assert.Equal(t, uPVal26, actualVal26) &&
		assert.True(t, inserted) &&
		assert.Equal(t, 26, *((*int)(actualVal26))) &&
		assert.Equal(t, 3, m.Size())
	if !good {
		return
	}

	p3a := Point{3,3,3}
	val27 := 27
	oldVal26, inserted := m.LoadOrStore(p3a, unsafe.Pointer(&val27))
	good = assert.Equal(t, uPVal26, oldVal26) &&
		assert.False(t, inserted) &&
		assert.Equal(t, 26, *((*int)(oldVal26))) &&
		assert.Equal(t, 3, m.Size())
	if !good {
		return
	}

	o26, b26 := m.Get(p3a)
	good = assert.Equal(t, uPVal26, o26) &&
		assert.Equal(t, 26, *((*int)(o26))) &&
		assert.Equal(t, true, b26)
	if !good {
		return
	}

	/*
		m.Clear()
		assert.Equal(t, 0, m.Size())
		o5, b5 := m.Get(&p1)
		assert.Equal(t, nil, o5)
		assert.Equal(t, false, b5)
	*/
}

func TestPointMapConflicts(t *testing.T) {
	// fist run on small init size, than big amount, than big init size
	good := runPointMapConflicts(t, 25, 2, 12, CInt(3)) &&
		runPointMapConflicts(t, 25*5, 8, 5, CInt(5)) &&
		runPointMapConflicts(t, 5*3*3*3, 8, 5, CInt(3))
	if !good {
		return
	}

}

func runPointMapConflicts(t *testing.T, hashSize, nbParallelProc, maxConflicts int, rdMax CInt) bool {
	m := MakePointHashMap(hashSize)
	//m.SetMaxConflictsAllowed(maxConflicts)

	nilPointer := unsafe.Pointer((*DInt)(nil))
	currentSize := 0
	for x := -rdMax; x <= rdMax; x++ {
		for y := -rdMax; y <= rdMax; y++ {
			for z := -rdMax; z <= rdMax; z++ {
				if !assert.Equal(t, currentSize, m.Size()) {
					return false
				}

				p := Point{x, y, z}
				ds := p.DistanceSquared()
				o := m.Put(p, unsafe.Pointer(&ds))
				if !assert.Equal(t, nilPointer, o) {
					return false
				}

				currentSize++
				if !assert.Equal(t, currentSize, m.Size()) {
					return false
				}
			}
		}
	}
	phm, ok := m.(*PointHashMap)
	assert.True(t, ok)
	//assert.True(t, phm.showedError)
	Log.Infof("Map size=%d with maxConflicts=%d", m.Size(), phm.concMap.mHashConflicts)

	Log.SetTrace()
	testMap := make(map[Point]*int32, m.Size())
	rc := MakeRangeContext(false, 1, Log)
	// First fill the map with one proc with value 1
	m.Range(func(point Point, value unsafe.Pointer) bool {
		already, exists := testMap[point]
		if !assert.False(t, exists, "Received %v %v twice", point, already) {
			return true
		}
		val1 := int32(1)
		testMap[point] = &val1
		return false
	}, rc)
	good := assert.Equal(t, m.Size(), len(testMap))
	if !good {
		return false
	}

	rc = MakeRangeContext(true, nbParallelProc, Log)
	// Then concurrently test passing only once
	m.Range(func(point Point, value unsafe.Pointer) bool {
		already, exists := testMap[point]
		if !assert.True(t, exists, "Should have get %v in map", point) {
			return true
		}
		if !assert.Equal(t, int32(1), *already, "Received wrong value for %v %v", point, already) {
			return true
		}
		atomic.AddInt32(already, 1)
		return false
	}, rc)
	rc.Wait()
	good = assert.Equal(t, m.Size(), len(testMap))
	if !good {
		return false
	}

	return true
}

func TestPointMapConcurrency(t *testing.T) {
	runConcurrencyTest(t, CInt(5), 4, 150, 6, 0.5, false, false)
	runConcurrencyTest(t, CInt(3), 2, 100, 6, 0.2, true, false)
}

func TestPointMapLoadOrStore(t *testing.T) {
	runConcurrencyTest(t, CInt(5), 4, 150, 6, 0.5, false, true)
	runConcurrencyTest(t, CInt(3), 2, 100, 6, 0.2, true, true)
}

type wasHereList struct {
	mutex  sync.Mutex
	runIds []int
}

func (whl *wasHereList) add(id int) {
	whl.mutex.Lock()
	defer whl.mutex.Unlock()
	if !whl.has(id) {
		whl.runIds = append(whl.runIds, id)
	}
}

func (whl *wasHereList) has(id int) bool {
	for _, i := range whl.runIds {
		if i == id {
			return true
		}
	}
	return false
}

func runConcurrencyTest(t *testing.T, rdMax CInt, divider, nbRoutines, maxConflicts int, hashSizeRatio float64, shouldMaxConflict bool, loadAndStore bool) bool {
	// First create a large collection of points
	rangeC := int(rdMax + 1 + rdMax) // adding the neg numbers
	testSet := make([]Point, rangeC*rangeC*rangeC)
	idx := 0
	for x := -rdMax; x <= rdMax; x++ {
		for y := -rdMax; y <= rdMax; y++ {
			for z := -rdMax; z <= rdMax; z++ {
				testSet[idx] = Point{x, y, z}
				idx++
			}
		}
	}

	if !assert.Equal(t, len(testSet), idx) {
		return false
	}
	dataSetSize := len(testSet)

	m := MakePointHashMap(int(float64(dataSetSize) * hashSizeRatio))
	//m.SetMaxConflictsAllowed(maxConflicts)
	//m.(*PointHashMap).fullLocked = false

	nbRound := int(dataSetSize/divider) + divider - 1
	good := assert.True(t, nbRoutines*nbRound > divider*dataSetSize, "not enough data %d x %d with nbRoutines=%d and nbRound=%d", dataSetSize, divider, nbRoutines, nbRound)
	// Enough concurrency
	good = good && assert.True(t, float64(nbRoutines)/float64(divider) > 16.0, "not enough concurrency with %d and %d", nbRoutines, divider)
	if !good {
		return false
	}

	failed := false
	start := time.Now()
	wg := new(sync.WaitGroup)
	for r := 0; r < nbRoutines; r++ {
		offset := (r % divider) * (dataSetSize / divider)
		runId := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < nbRound && !failed; i++ {
				idx := offset + i
				if idx >= dataSetSize {
					idx = dataSetSize - 1
				}
				if loadAndStore {
					whlI, _ := m.LoadOrStore(testSet[idx], unsafe.Pointer(new(wasHereList)))
					whl := (*wasHereList)(whlI)
					if whl == nil {
						failed = true
						return
					}
					whl.add(runId)
				} else {
					m.Put(testSet[idx], unsafe.Pointer(&idx))
				}
			}
		}()
	}
	wg.Wait()
	if failed {
		return assert.Fail(t, "failed load and store")
	}
	Log.Infof("It took %v to put %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
	phm, ok := m.(*PointHashMap)
	Log.Infof("Map size=%d with maxConflicts=%d", m.Size(), phm.concMap.mHashConflicts)
	assert.True(t, ok)
	//assert.Equal(t, shouldMaxConflict, phm.showedError)
	if !assert.Equal(t, dataSetSize, m.Size()) {
		return false
	}

	if loadAndStore {
		start := time.Now()
		wg := new(sync.WaitGroup)
		wg.Add(nbRoutines)
		for r := 0; r < nbRoutines; r++ {
			offset := (r % divider) * (dataSetSize / divider)
			runId := r
			go func() {
				defer wg.Done()
				for i := 0; i < nbRound && !failed; i++ {
					idx := offset + i
					if idx >= dataSetSize {
						idx = dataSetSize - 1
					}
					whl, b := m.Get(testSet[idx])
					whList := (*wasHereList)(whl)
					good := assert.True(t, b) && assert.NotNil(t, whList) &&
						assert.True(t, whList.has(runId), "point %v of %d / %d / %d failed to have %d in %v", testSet[idx], nbRoutines, divider, nbRound, r, whList.runIds)
					if !good {
						failed = true
						return
					}
				}
			}()
		}
		wg.Wait()
		Log.Infof("It took %v to test %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
	}
	return true
}
