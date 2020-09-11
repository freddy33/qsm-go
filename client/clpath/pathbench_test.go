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
	client.Log.SetWarn()
	client.Log.SetAssert(true)
	Log.SetWarn()
	Log.SetAssert(true)
	m3util.SetToTestMode()

	env := client.GetFullApiTestEnv(m3util.PathTestEnv)
	pointData := client.GetClientPointPackData(env)
	pathData := client.GetClientPathPackData(env)

	for r := 0; r < N; r++ {
		//		for _, ctx := range allCtx[pType] {
		start := time.Now()
		growthCtx := pointData.GetGrowthContextByTypeAndIndex(pType, 0)
		pathCtx := pathData.CreatePathCtxFromAttributes(growthCtx, 0, m3point.Origin)
		runPathContext(pathCtx, until)
		t := time.Since(start)
		LogDataTest.Infof("%s %s %d %d", t, pathCtx, pathCtx.CountAllPathNodes(), pathCtx.GetNumberOfOpenNodes())
		//			if single {
		//				break
		//			}
		//		}
	}
}

func runPathContext(pathCtx m3path.PathContext, until int) {
	pathCtx.InitRootNode(m3point.Origin)
	for d := 0; d < until; d++ {
		verifyDistance(pathCtx, d)
		origLen := float64(pathCtx.GetNumberOfOpenNodes())
		predictedIntLen := pathCtx.PredictedNextOpenNodesLen()
		pathCtx.MoveToNextNodes()
		if d != 0 && LogDataTest.IsInfo() {
			finalLen := pathCtx.GetNumberOfOpenNodes()
			df := float64(d)
			predictedRatio := 1.0 + 2.0/df + 1.0/(df*df)
			actualRatio := float64(finalLen) / origLen
			errorBar := math.Abs(actualRatio-predictedRatio) / predictedRatio
			if predictedIntLen < finalLen {
				LogDataTest.Errorf("%s: Distance %d orig=%.0f final=%.0f predictLen=%d actualRatio=%.5f predictedRatio=%.5f errorBar=%f", pathCtx.String(), d, origLen, finalLen, predictedIntLen, actualRatio, predictedRatio, errorBar)
			} else {
				LogDataTest.Infof("%s: Distance %d orig=%.0f final=%.0f predictLen=%d actualRatio=%.5f predictedRatio=%.5f errorBar=%f", pathCtx.String(), d, origLen, finalLen, predictedIntLen, actualRatio, predictedRatio, errorBar)
			}
		}
	}
}

func verifyDistance(pathCtx m3path.PathContext, d int) {
	pnm := pathCtx.GetPathNodeMap()
	pnm.Range(func(point m3point.Point, pn m3path.PathNode) bool {
		if pn.D() != d {
			client.Log.Errorf("Something fishy for %s", pathCtx.String())
			return true
		}
		return false
	}, 1)
}
