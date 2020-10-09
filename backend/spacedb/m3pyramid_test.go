package spacedb

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var LogData = m3util.NewDataLogger("m3data", m3util.INFO)

func TestCreateAllIndexes(t *testing.T) {
	allContexts := m3point.GetAllGrowthTypes()
	for _, ctxType := range allContexts {
		createAllIndexesForContext(t, ctxType)
	}
}

func createAllIndexesForContext(t assert.TestingT, ctxType m3point.GrowthType) [][4]int {
	nbIndexes := ctxType.GetNbIndexes()
	res, idxs := m3space.CreateAllIndexes(nbIndexes)
	assert.NotNil(t, res)
	for i := 0; i < len(idxs)/2; i++ {
		assert.Equal(t, idxs[i*2], idxs[i*2+1], "failed index value for %d %v", i, ctxType)
	}
	return res
}

var envMutex sync.Mutex
var spaceEnv m3util.QsmEnvironment

func getSpaceTestEnv() m3util.QsmEnvironment {
	if spaceEnv != nil {
		return spaceEnv
	}

	envMutex.Lock()
	defer envMutex.Unlock()
	if spaceEnv != nil {
		return spaceEnv
	}
	m3util.SetToTestMode()
	spaceEnv := GetSpaceDbCleanEnv(m3util.SpaceTempEnv)
	return spaceEnv
}

func TestSpaceAllPyramids(t *testing.T) {
	LogData.SetInfo()

	Log.SetWarn()
	m3path.Log.SetWarn()
	LogStat.SetWarn()
	LogRun.SetWarn()

	ctxs := [4]m3point.GrowthType{8, 8, 8, 8}

	// TODO: go through the offsets
	offsets := [4]int{0, 0, 0, 0}
	doPercent := float32(0.002)

	maxSize := m3point.CInt(6)
	minSize := m3point.CInt(5)
	LogData.Infof("Going from min %d to max %d pyramid size", minSize, maxSize)
	for pSize := minSize; pSize <= maxSize; pSize++ {
		nbFound := 0
		allIndexes := createAllIndexesForContext(t, 8)
		LogData.Infof("Running pyramid check for size %d and %d indexes", pSize, len(allIndexes))
		for i, idxs := range allIndexes {
			spaceName := fmt.Sprintf("TestSpaceAllPyramids-%d-%d", pSize, i)
			rf := rand.Float32()
			Log.Debugf("Comparing %f < %f for %s indexes = %v", rf, doPercent, spaceName, idxs)
			if rf > doPercent {
				continue
			}
			start := time.Now()
			LogData.Infof("Running space %s indexes = %v", spaceName, idxs)
			space := createNewSpace(t, spaceName, m3space.ZeroDistAndTime)
			found, originalPyramid, foundTime, finalPyramid, nbPoss := RunSpacePyramidWithParams(space, pSize, ctxs, idxs, offsets)
			if found {
				orgSize := GetPyramidSize(originalPyramid)
				finalSize := GetPyramidSize(finalPyramid)
				diff := m3point.AbsDInt(orgSize - finalSize)
				ratio := float64(diff) / float64(orgSize)
				LogData.Infof("%d %d %v %d %d %d %d %d %.5f",
					pSize, 8, idxs, foundTime, nbPoss, orgSize, finalSize, diff, ratio)
				nbFound++
			}
			LogData.Infof("Running space %s until %d took %v", spaceName, space.GetMaxTime(), time.Now().Sub(start))
			if nbFound > 3 {
				break
			}
		}
	}
}

func TestSpaceRunPySize5(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	runSpaceTest(t, 5, "TestSpaceRunPySize5")
}

func TestSpaceRunPySize4(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()

	space := createNewSpace(t, "TestSpaceRunPySize4-1", m3space.ZeroDistAndTime)

	found, originalPyramid, time, finalPyramid, nbPoss := RunSpacePyramidWithParams(space, 4, [4]m3point.GrowthType{2, 2, 2, 2}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	// TODO: Reactivate after space node fix
	//assert.True(t, found)
	orgSize := GetPyramidSize(originalPyramid)
	finalSize := GetPyramidSize(finalPyramid)
	diff := m3point.AbsDInt(orgSize - finalSize)
	LogStat.Infof("%v %d %v %v %d %d %d %d", found, time, originalPyramid, finalPyramid, nbPoss, orgSize, finalSize, diff)

	space = createNewSpace(t, "TestSpaceRunPySize4-2", m3space.ZeroDistAndTime)

	found, originalPyramid, time, finalPyramid, nbPoss = RunSpacePyramidWithParams(space, 4, [4]m3point.GrowthType{2, 2, 2, 2}, [4]int{0, 0, 0, 3}, [4]int{0, 0, 0, 0})
	// TODO: Reactivate after space node fix
	//assert.True(t, found)
	orgSize = GetPyramidSize(originalPyramid)
	finalSize = GetPyramidSize(finalPyramid)
	diff = m3point.AbsDInt(orgSize - finalSize)
	LogStat.Infof("%v %d %v %v %d %d %d %d", found, time, originalPyramid, finalPyramid, nbPoss, orgSize, finalSize, diff)
}

func TestSpaceRunPySize3(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	runSpaceTest(t, 3, "TestSpaceRunPySize3")
}

func TestSpaceRunPySize2(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	runSpaceTest(t, 2, "TestSpaceRunPySize2")
}

func runSpaceTest(t *testing.T, pSize m3point.CInt, spaceName string) {
	space := createNewSpace(t, spaceName, m3space.ZeroDistAndTime)
	growthTypes := [4]m3point.GrowthType{8, 8, 8, 8}
	indexes := [4]int{0, 4, 8, 10}
	offsets := [4]int{0, 0, 0, 4}
	RunSpacePyramidWithParams(space, pSize, growthTypes, indexes, offsets)
}
