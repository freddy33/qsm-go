package clpath

import (
	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
	"time"
)

var LogDataTest = m3util.NewDataLogger("CLDATA", m3util.DEBUG)
var Log = m3util.NewLogger("clpath", m3util.INFO)

/***************************************************************/
// PathContext Test size optimization
/***************************************************************/

func TestClientAllPathCtx(t *testing.T) {
	LogDataTest.SetInfo()
	client.Log.SetInfo()
	client.Log.SetAssert(true)
	Log.SetInfo()
	Log.SetAssert(true)
	m3util.SetToTestMode()

	env := client.GetInitializedApiEnv(m3util.TestClientEnv)

	for _, growthType := range m3point.GetAllGrowthTypes() {
		if !runForPathCtxType(t, env, 25, growthType, 0.1) {
			return
		}
	}
}

func runForPathCtxType(t *testing.T, env *client.QsmApiEnvironment, until int, pType m3point.GrowthType, doPercent float32) bool {
	pathData := client.GetClientPathPackData(env)

	nbIndexes := pType.GetNbIndexes()
	maxOffset := pType.GetMaxOffset()

	for idx := 0; idx < nbIndexes; idx ++ {
		for offset := 0; offset < maxOffset; offset++ {
			rf := rand.Float32()
			Log.Infof("Comparing %f < %f for %d-%d-%d", rf, doPercent, pType, idx, offset)
			if rf < doPercent {
				start := time.Now()
				pathCtx, _ := pathData.GetPathCtxFromAttributes(pType, idx, offset)
				allNb, lastNb, err := runPathContext(pathCtx, until)
				if !assert.NoError(t, err) {
					return false
				}
				timeTook := time.Since(start)
				LogDataTest.Infof("%s %s %d %d", timeTook, pathCtx, allNb, lastNb)
			}
		}
	}

	return true
}

func runPathContext(pathCtx m3path.PathContext, until int) (int, int, error) {
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
				LogDataTest.Infof("%s: Distance %d finalLen=%d predictLen=%d errorBar=%f", pathCtx.String(), d, finalLen, predictedIntLen, errorBar)
			}
		}
	}
	return pathCtx.GetNumberOfNodesBetween(0, until), pathCtx.GetNumberOfNodesAt(until), nil
}
