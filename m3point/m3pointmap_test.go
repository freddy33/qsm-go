package m3point

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPointMapBasic(t *testing.T) {
	m := MakePointHashMap(10)
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
}

func TestPointMapConflicts(t *testing.T) {
	m := MakePointHashMap(4)
	m.SetMaxConflictsAllowed(25)

	currentSize := 0
	for x := CInt(-3); x < 4; x++ {
		for y := CInt(-3); y < 4; y++ {
			for z := CInt(-3); z < 4; z++ {
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

	dataSetSize := len(testSet)

	m := MakePointHashMap(int(float64(dataSetSize) * hashSizeRatio))
	m.SetMaxConflictsAllowed(maxConflicts)
	m.(*pointHashMap).fullLocked = false

	nbRound := int(dataSetSize/divider) + divider - 1
	assert.True(t, nbRoutines*nbRound > divider*dataSetSize, "not enough data %d x %d with nbRoutines=%d and nbRound=%d", dataSetSize, divider, nbRoutines, nbRound)
	// Enough concurrency
	assert.True(t, float64(nbRoutines)/float64(divider) > 16.0, "not enough concurrency with %d and %d", nbRoutines, divider)

	start := time.Now()
	wg := new(sync.WaitGroup)
	for r := 0; r < nbRoutines; r++ {
		offset := (r % divider) * (dataSetSize / divider)
		wg.Add(1)
		go func() {
			for i := 0; i < nbRound; i++ {
				idx := offset + i
				if idx >= dataSetSize {
					idx = dataSetSize - 1
				}
				if loadAndStore {
					whl, _ := m.LoadOrStore(&testSet[idx], new(wasHereList))
					whl.(*wasHereList).add(r)
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
		for r := 0; r < nbRoutines; r++ {
			offset := (r % divider) * (dataSetSize / divider)
			wg.Add(1)
			go func() {
				for i := 0; i < nbRound; i++ {
					idx := offset + i
					if idx >= dataSetSize {
						idx = dataSetSize - 1
					}
					whl, b := m.Get(&testSet[idx])
					assert.True(t, b)
					whList := whl.(*wasHereList)
					assert.True(t, whList.has(r), "point %v failed to have %d in %v", testSet[idx], r, whList.runIds)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		Log.Infof("It took %v to test %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
	}
}
