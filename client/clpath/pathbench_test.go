package clpath

import (
	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"math"
	"testing"
	"time"
)

var LogDataTest = m3util.NewDataLogger("DATA", m3util.DEBUG)
var Log = m3util.NewLogger("clpath", m3util.INFO)

const (
	BenchNbRound = 51
)

/***************************************************************/
// PathContext Test size optimization
/***************************************************************/

func TestClientPathCtx8(t *testing.T) {
	runForPathCtxType(1, 25, m3point.GrowthType(8), true)
}

/***************************************************************/
// PathContext Bench
/***************************************************************/

func BenchmarkPathCtx3(b *testing.B) {
	runForPathCtxType(b.N, BenchNbRound, 3, false)
}

func BenchmarkPathCtx4(b *testing.B) {
	runForPathCtxType(b.N, BenchNbRound, 4, false)
}

func BenchmarkPathCtx8(b *testing.B) {
	runForPathCtxType(b.N, BenchNbRound, 8, false)
}

func runForPathCtxType(N, until int, pType m3point.GrowthType, single bool) {
	LogDataTest.SetInfo()
	client.Log.SetInfo()
	client.Log.SetAssert(true)
	Log.SetInfo()
	Log.SetAssert(true)
	m3util.SetToTestMode()

	env := client.GetInitializedApiEnv(m3util.PathTestEnv)
	pathData := client.GetClientPathPackData(env)

	for r := 0; r < N; r++ {
		//		for _, ctx := range allCtx[pType] {
		start := time.Now()
		pathCtx, _ := pathData.GetPathCtxFromAttributes(pType, 0, 0)
		runPathContext(pathCtx, until)
		t := time.Since(start)
		LogDataTest.Infof("%s %s %d %d", t, pathCtx, pathCtx.GetNumberOfNodesBetween(0, until), pathCtx.GetNumberOfNodesAt(until))
		//			if single {
		//				break
		//			}
		//		}
	}
}

func runPathContext(pathCtx m3path.PathContext, until int) {
	for d := 0; d < until; d++ {
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
