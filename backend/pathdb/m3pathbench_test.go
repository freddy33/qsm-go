package pathdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
	"time"
)

var LogDataTest = m3util.NewDataLogger("DATA", m3util.DEBUG)

/***************************************************************/
// PathContext Test size optimization
/***************************************************************/

func TestPopulateMaxAllPathCtx(t *testing.T) {
	LogDataTest.SetInfo()
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetWarn()
	m3point.Log.SetAssert(true)
	m3util.SetToTestMode()

	env := GetPathDbFullEnv(m3util.PathTestEnv)
	for _, growthType := range m3point.GetAllGrowthTypes() {
		runForPathCtxType(t, env, 25, growthType, 0.1)
	}
}

func runForPathCtxType(t *testing.T, env *m3db.QsmDbEnvironment, until int, growthType m3point.GrowthType, doPercent float32) {
	pathData := GetServerPathPackData(env)
	err := pathData.initAllPathContexts()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	for _, pathCtx := range pathData.AllCenterContexts[growthType] {
		rf := rand.Float32()
		Log.Infof("Comparing %f < %f for %s", rf, doPercent, pathCtx.String())
		if rf < doPercent {
			start := time.Now()
			allNb, lastNb, err := runPathContext(pathCtx, until)
			assert.NoError(t, err)
			t := time.Since(start)
			LogDataTest.Infof("%s %s %d %d", t, pathCtx, allNb, lastNb)
		}
	}
}

func runPathContext(pathCtx *PathContextDb, until int) (int, int, error) {
	for d := 0; d < until; d++ {
		maxDist := pathCtx.GetMaxDist()
		if d > maxDist {
			err := pathCtx.calculateNextMaxDist()
			if err != nil {
				return -1, -1, err
			}
		}
		if LogDataTest.IsInfo() {
			predictedIntLen := m3path.CalculatePredictedSize(pathCtx.GetGrowthType(), d)
			finalLen := pathCtx.GetNumberOfNodesAt(d)
			errorBar := math.Abs(float64(finalLen-predictedIntLen)) / float64(finalLen)
			if predictedIntLen < finalLen {
				LogDataTest.Errorf("%s: Distance %d finalLen=%d predictLen=%d errorBar=%f", pathCtx.String(), d, finalLen, predictedIntLen, errorBar)
			} else {
				LogDataTest.Infof("%s: Distance %d finalLen=%d predictLen=%d errorBar=%f", pathCtx.String(), d, finalLen, predictedIntLen, errorBar)
			}
		}
	}
	return pathCtx.GetNumberOfNodesBetween(0, until), pathCtx.GetNumberOfNodesAt(until), nil
}
