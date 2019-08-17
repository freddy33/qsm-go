package m3point

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPointMapBasic(t *testing.T) {
	m := MakePointHashMap(10, 2)
	p1 := Point{1, 2, 3}
	o1, b1 := m.Put(&p1, 23)
	assert.Equal(t, nil, o1)
	assert.Equal(t, true, b1)
	assert.Equal(t, 1, m.Size())
	p2 := Point{1, 2, 3}
	o2, b2 := m.Get(&p2)
	assert.Equal(t, 23, o2)
	assert.Equal(t, true, b2)
	p2[0] = 2
	o3, b3 := m.Get(&p2)
	assert.Equal(t, nil, o3)
	assert.Equal(t, false, b3)
	assert.Equal(t, 1, m.Size())
	o4, b4 := m.Put(&p2, 24)
	assert.Equal(t, nil, o4)
	assert.Equal(t, true, b4)
	assert.Equal(t, 2, m.Size())

	m.Clear()
	assert.Equal(t, 0, m.Size())
	o5, b5 := m.Get(&p1)
	assert.Equal(t, nil, o5)
	assert.Equal(t, false, b5)
}

func TestPointMapConflicts(t *testing.T) {
	runPointMapConflicts(t, 25, 2, 12, CInt(3))
	runPointMapConflicts(t, 25*5, 8, 5, CInt(5))
}

func runPointMapConflicts(t *testing.T, hashSize, nbSegments, maxConflicts int, rdMax CInt) {
	m := MakePointHashMap(hashSize, nbSegments)
	m.SetMaxConflictsAllowed(maxConflicts)

	currentSize := 0
	for x := -rdMax; x <= rdMax; x++ {
		for y := -rdMax; y <= rdMax; y++ {
			for z := -rdMax; z <= rdMax; z++ {
				assert.Equal(t, currentSize, m.Size())

				p := Point{x, y, z}
				o, b := m.Put(&p, p.DistanceSquared())
				assert.Equal(t, nil, o)
				assert.Equal(t, true, b)

				currentSize++
				assert.Equal(t, currentSize, m.Size())
			}
		}
	}
	phm, ok := m.(*pointHashMap)
	assert.True(t, ok)
	assert.True(t, phm.showedError)
	Log.Infof("Map size=%d with maxConflicts=%d", m.Size(), m.GetCurrentMaxConflicts())

	testMap := make(map[Point]DInt, m.Size())
	m.Range(func(point Point, value interface{}) bool {
		already, exists := testMap[point]
		assert.False(t, exists, "Received %v %v twice", point, already)
		testMap[point] = value.(DInt)
		return false
	})
	assert.Equal(t, m.Size(), len(testMap))
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

func runConcurrencyTest(t *testing.T, rdMax CInt, divider, nbRoutines, maxConflicts int, hashSizeRatio float64, shouldMaxConflict bool, loadAndStore bool) {
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

	assert.Equal(t, len(testSet), idx)
	dataSetSize := len(testSet)

	m := MakePointHashMap(int(float64(dataSetSize)*hashSizeRatio), 16)
	m.SetMaxConflictsAllowed(maxConflicts)
	m.(*pointHashMap).fullLocked = false

	nbRound := int(dataSetSize/divider) + divider - 1
	assert.True(t, nbRoutines*nbRound > divider*dataSetSize, "not enough data %d x %d with nbRoutines=%d and nbRound=%d", dataSetSize, divider, nbRoutines, nbRound)
	// Enough concurrency
	assert.True(t, float64(nbRoutines)/float64(divider) > 16.0, "not enough concurrency with %d and %d", nbRoutines, divider)

	start := time.Now()
	wg := new(sync.WaitGroup)
	wg.Add(nbRoutines)
	for r := 0; r < nbRoutines; r++ {
		offset := (r % divider) * (dataSetSize / divider)
		runId := r
		go func() {
			for i := 0; i < nbRound; i++ {
				idx := offset + i
				if idx >= dataSetSize {
					idx = dataSetSize - 1
				}
				if loadAndStore {
					whl, _ := m.LoadOrStore(&testSet[idx], new(wasHereList))
					whl.(*wasHereList).add(runId)
				} else {
					m.Put(&testSet[idx], idx)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	Log.Infof("It took %v to put %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
	Log.Infof("Map size=%d with maxConflicts=%d", m.Size(), m.GetCurrentMaxConflicts())
	phm, ok := m.(*pointHashMap)
	assert.True(t, ok)
	assert.Equal(t, shouldMaxConflict, phm.showedError)
	assert.Equal(t, dataSetSize, m.Size())

	if loadAndStore {
		start := time.Now()
		wg := new(sync.WaitGroup)
		wg.Add(nbRoutines)
		for r := 0; r < nbRoutines; r++ {
			offset := (r % divider) * (dataSetSize / divider)
			runId := r
			go func() {
				for i := 0; i < nbRound; i++ {
					idx := offset + i
					if idx >= dataSetSize {
						idx = dataSetSize - 1
					}
					whl, b := m.Get(&testSet[idx])
					assert.True(t, b)
					whList := whl.(*wasHereList)
					assert.True(t, whList.has(runId), "point %v of %d / %d / %d failed to have %d in %v", testSet[idx], nbRoutines, divider, nbRound, r, whList.runIds)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		Log.Infof("It took %v to test %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
	}
}
