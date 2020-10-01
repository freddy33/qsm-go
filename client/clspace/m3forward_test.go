package spacedb

import (
	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var Log = m3util.NewLogger("clspace", m3util.INFO)
var LogData = m3util.NewDataLogger("m3data", m3util.INFO)

func BenchmarkPack1(b *testing.B) {
	benchSpaceTest(b, 1)
}

func BenchmarkPack2(b *testing.B) {
	benchSpaceTest(b, 2)
}

func BenchmarkPack12(b *testing.B) {
	benchSpaceTest(b, 12)
}

func BenchmarkPack20(b *testing.B) {
	benchSpaceTest(b, 20)
}

func benchSpaceTest(b *testing.B, pSize m3point.CInt) {
	Log.SetWarn()
	m3space.LogStat.SetWarn()
	m3space.LogRun.SetWarn()
	for r := 0; r < b.N; r++ {
		runSpaceTest(pSize)
	}
}

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
	envMutex.Lock()
	defer envMutex.Unlock()
	if spaceEnv != nil {
		return spaceEnv
	}
	m3util.SetToTestMode()
	spaceEnv := client.GetInitializedApiEnv(m3util.TestClientEnv)
	return spaceEnv
}

func TestSpaceAllPyramids(t *testing.T) {
	LogData.SetInfo()
	m3path.Log.SetWarn()
	Log.SetWarn()
	m3space.LogStat.SetWarn()
	m3space.LogRun.SetWarn()

	env := getSpaceTestEnv()

	ctxs := [4]m3point.GrowthType{8, 8, 8, 8}

	// TODO: go through the offsets
	offsets := [4]int{0, 0, 0, 0}

	LogData.Infof("Size Type Idxs time nbPoss orgSize finalSize diff ratio")
	maxSize := m3point.CInt(10)
	maxIndexes := 20
	for pSize := m3point.CInt(8); pSize <= maxSize; pSize++ {
		nbFound := 0
		allIndexes := createAllIndexesForContext(t, 8)
		for i, idxs := range allIndexes {
			found, originalPyramid, time, finalPyramid, nbPoss := m3space.RunSpacePyramidWithParams(env, pSize, ctxs, idxs, offsets)
			if found {
				orgSize := m3space.GetPyramidSize(originalPyramid)
				finalSize := m3space.GetPyramidSize(finalPyramid)
				diff := m3point.AbsDInt(orgSize - finalSize)
				ratio := float64(diff) / float64(orgSize)
				LogData.Infof("%d %d %v %d %d %d %d %d %.5f",
					pSize, 8, idxs, time, nbPoss, orgSize, finalSize, diff, ratio)
				nbFound++
			}
			if nbFound > 10 || i > maxIndexes {
				break
			}
		}
	}
}

func TestSpaceRunPySize5(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()
	runSpaceTest(5)
}

func TestSpaceRunPySize4(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()

	env := getSpaceTestEnv()

	found, originalPyramid, time, finalPyramid, nbPoss := m3space.RunSpacePyramidWithParams(env, 4, [4]m3point.GrowthType{2, 2, 2, 2}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	// TODO: Reactivate after space node fix
	//assert.True(t, found)
	orgSize := m3space.GetPyramidSize(originalPyramid)
	finalSize := m3space.GetPyramidSize(finalPyramid)
	diff := m3point.AbsDInt(orgSize - finalSize)
	m3space.LogStat.Infof("%v %d %v %v %d %d %d %d", found, time, originalPyramid, finalPyramid, nbPoss, orgSize, finalSize, diff)

	found, originalPyramid, time, finalPyramid, nbPoss = m3space.RunSpacePyramidWithParams(env, 4, [4]m3point.GrowthType{2, 2, 2, 2}, [4]int{0, 0, 0, 3}, [4]int{0, 0, 0, 0})
	// TODO: Reactivate after space node fix
	//assert.True(t, found)
	orgSize = m3space.GetPyramidSize(originalPyramid)
	finalSize = m3space.GetPyramidSize(finalPyramid)
	diff = m3point.AbsDInt(orgSize - finalSize)
	m3space.LogStat.Infof("%v %d %v %v %d %d %d %d", found, time, originalPyramid, finalPyramid, nbPoss, orgSize, finalSize, diff)
}

func TestSpaceRunPySize3(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()
	runSpaceTest(3)
}

func TestSpaceRunPySize2(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()
	runSpaceTest(2)
}

func runSpaceTest(pSize m3point.CInt) {
	env := getSpaceTestEnv()
	growthTypes := [4]m3point.GrowthType{8, 8, 8, 8}
	indexes := [4]int{0, 4, 8, 10}
	offsets := [4]int{0, 0, 0, 4}
	m3space.RunSpacePyramidWithParams(env, pSize, growthTypes, indexes, offsets)
}
