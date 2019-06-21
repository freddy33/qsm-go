package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/gonum/stat"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func benchSpaceTest(b *testing.B, pSize int64) {
	Log.SetWarn()
	LogStat.SetWarn()
	LogRun.SetWarn()
	for r := 0; r < b.N; r++ {
		runSpaceTest(pSize)
	}
}

func TestCreateAllIndexes(t *testing.T) {
	allContexts := m3point.GetAllContextTypes()
	for _, ctxType := range allContexts {
		createAllIndexesForContext(t, ctxType)
	}
}

func createAllIndexesForContext(t assert.TestingT, ctxType m3point.ContextType) [][4]int {
	nbIndexes := ctxType.GetNbIndexes()
	res, idxs := createAllIndexes(nbIndexes)
	assert.NotNil(t, res)
	for i := 0; i < len(idxs)/2; i++ {
		assert.Equal(t, idxs[i*2], idxs[i*2+1], "failed index value for %d %v",i,ctxType)
	}
	return res
}

func TestSpaceAllPyramids(t *testing.T) {
	Log.SetWarn()
	LogStat.SetWarn()
	LogRun.SetWarn()
	allContexts := m3point.GetAllContextTypes()
	LogData.Infof("Size Type Idxs time nbPoss orgSize finalSize diff ratio")
	maxSize := int64(4)
	maxIndexes := 200
	for pSize := int64(4); pSize <= maxSize; pSize++ {
		for _, ctxType := range allContexts {
			nbFound := 0
			ctxs := [4]m3point.ContextType{ctxType, ctxType, ctxType, ctxType}
			allIndexes := createAllIndexesForContext(t, ctxType)
			for i, idxs := range allIndexes {
				found, originalPyramid, time, finalPyramid, nbPoss := runSpacePyramidWithParams(pSize, ctxs, idxs, [4]int{0, 0, 0, 0})
				if found {
					orgSize := GetPyramidSize(originalPyramid)
					finalSize := GetPyramidSize(finalPyramid)
					diff := m3point.Abs64(orgSize-finalSize)
					ratio := float64(diff) / float64(orgSize)
					LogData.Infof("%d %d %v %d %d %d %d %d %.5f",
						pSize, ctxType, idxs, time, nbPoss, orgSize, finalSize, diff, ratio)
					nbFound++
				}
				if nbFound > 10 || i > maxIndexes {
					break
				}
			}
		}
	}
}

func TestSpaceRunPySize5(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	runSpaceTest(5)
}

func TestSpaceRunPySize4(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	found, originalPyramid, time, finalPyramid, nbPoss := runSpacePyramidWithParams(4, [4]m3point.ContextType{2, 2, 2, 2}, [4]int{0, 0, 0, 0}, [4]int{0, 0, 0, 0})
	assert.True(t, found)
	orgSize := GetPyramidSize(originalPyramid)
	finalSize := GetPyramidSize(finalPyramid)
	diff := m3point.Abs64(orgSize-finalSize)
	LogStat.Infof("%d %v %v %d %d %d %d",time, originalPyramid, finalPyramid, nbPoss, orgSize, finalSize, diff)

	found, originalPyramid, time, finalPyramid, nbPoss = runSpacePyramidWithParams(4, [4]m3point.ContextType{2, 2, 2, 2}, [4]int{0, 0, 0, 3}, [4]int{0, 0, 0, 0})
	assert.True(t, found)
	orgSize = GetPyramidSize(originalPyramid)
	finalSize = GetPyramidSize(finalPyramid)
	diff = m3point.Abs64(orgSize-finalSize)
	LogStat.Infof("%d %v %v %d %d %d %d",time, originalPyramid, finalPyramid, nbPoss, orgSize, finalSize, diff)
}

func TestSpaceRunPySize3(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	runSpaceTest(3)
}

func TestSpaceRunPySize2(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	runSpaceTest(2)
}

func runSpaceTest(pSize int64) {
	runSpacePyramidWithParams(pSize, [4]m3point.ContextType{8, 8, 8, 8}, [4]int{0, 4, 8, 10}, [4]int{0, 0, 0, 4})
}


func TestStdDev(t *testing.T) {
	fmt.Println(stat.StdDev([]float64{1.3, 1.5, 1.7, 1.1}, nil))
}

