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
	LogDataTest.SetWarn()
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetWarn()
	m3point.Log.SetAssert(true)
	m3util.SetToTestMode()

	env := GetPathDbFullEnv(m3util.PathTestEnv)
	for _, growthType := range m3point.GetAllGrowthTypes() {
		if !runForPathCtxType(t, env, 25, growthType, 0.1) {
			return
		}
	}
}

func runForPathCtxType(t *testing.T, env *m3db.QsmDbEnvironment, until int, growthType m3point.GrowthType, doPercent float32) bool {
	pathData := GetServerPathPackData(env)
	err := pathData.InitAllPathContexts()
	if !assert.NoError(t, err) {
		return false
	}
	for _, pathCtx := range pathData.AllCenterContexts[growthType] {
		rf := rand.Float32()
		Log.Debugf("Comparing %f < %f for %s", rf, doPercent, pathCtx.String())
		if rf < doPercent {
			start := time.Now()
			allNb, lastNb, err := runPathContext(pathCtx, until)
			if !assert.NoError(t, err) {
				return false
			}
			timeTook := time.Since(start)
			LogDataTest.Infof("%s %s %d %d", timeTook, pathCtx, allNb, lastNb)
		}
	}
	return true
}

func runPathContext(pathCtx *PathContextDb, until int) (int, int, error) {
	err := pathCtx.RequestNewMaxDist(until)
	if err != nil {
		return -1, -1, err
	}
	for d := 0; d < until; d++ {
		if LogDataTest.IsInfo() {
			predictedIntLen := m3path.CalculatePredictedSize(pathCtx.GetGrowthType(), d)
			finalLen := pathCtx.GetNumberOfNodesAt(d)
			errorBar := math.Abs(float64(finalLen-predictedIntLen)) / float64(predictedIntLen)
			// If final length way too small => error
			if d > 10 && finalLen < predictedIntLen && errorBar > 0.3 {
				return -1, -1, m3util.MakeQsmErrorf("%s: Distance %d finalLen=%d predictLen=%d errorBar=%f", pathCtx.String(), d, finalLen, predictedIntLen, errorBar)
			}
			if predictedIntLen < finalLen && errorBar > 0.08 {
				LogDataTest.Errorf("%s: Distance %d finalLen=%d predictLen=%d errorBar=%f", pathCtx.String(), d, finalLen, predictedIntLen, errorBar)
			} else {
				LogDataTest.Debugf("%s: Distance %d finalLen=%d predictLen=%d errorBar=%f", pathCtx.String(), d, finalLen, predictedIntLen, errorBar)
			}
		}
	}
	return pathCtx.GetNumberOfNodesBetween(0, until), pathCtx.GetNumberOfNodesAt(until), nil
}
