package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"math"
	"testing"
	"time"
)

var LogDataTest = m3util.NewDataLogger("DATA", m3util.DEBUG)

const (
	BenchNbRound    = 51
)

/***************************************************************/
// PathContext Test size optimization
/***************************************************************/

func TestAllPathSize(t *testing.T) {
	for _, ctxType := range m3point.GetAllContextTypes() {
		runForPathCtxType(1, 128, ctxType, true)
	}
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

func runForPathCtxType(N, until int, pType m3point.ContextType, single bool) {
	LogDataTest.SetWarn()
	Log.SetWarn()
	Log.SetAssert(true)
	m3point.Log.SetWarn()
	m3point.Log.SetAssert(true)

	allCtx := getAllTestContexts()
	for r := 0; r < N; r++ {
		for _, ctx := range allCtx[pType] {
			start := time.Now()
			pathCtx := MakePathContext(ctx.GetType(), ctx.GetIndex(), ctx.offset, MakeSimplePathNodeMap(5*until*until))
			runPathContext(pathCtx, until)
			t := time.Since(start)
			LogDataTest.Infof("%s %s %d %d %d", t, pathCtx, pathCtx.GetPathNodeMap().GetSize(), len(pathCtx.openEndNodes), pathCtx.openEndNodes[0].pn.D())
			if single {
				break
			}
		}
	}
}

func runPathContext(pathCtx *PathContext, until int) {
	pathCtx.InitRootNode(m3point.Origin)
	for d := 0; d < until; d++ {
		verifyDistance(pathCtx, d)
		origLen := float64(len(pathCtx.openEndNodes))
		predictedIntLen := pathCtx.GetNextOpenNodesLen()
		pathCtx.MoveToNextNodes()
		if d != 0 && LogDataTest.IsInfo() {
			finalLen := len(pathCtx.openEndNodes)
			df := float64(d)
			predictedRatio := 1.0 + 2.0/df + 1.0/(df*df)
			actualRatio := float64(finalLen)/origLen
			errorBar := math.Abs(actualRatio-predictedRatio) / predictedRatio
			if predictedIntLen < finalLen {
				LogDataTest.Errorf("%s: Distance %d orig=%.0f final=%.0f predictLen=%d actualRatio=%.5f predictedRatio=%.5f errorBar=%f", pathCtx.String(), d, origLen, finalLen, predictedIntLen, actualRatio, predictedRatio, errorBar)
			} else {
				LogDataTest.Infof("%s: Distance %d orig=%.0f final=%.0f predictLen=%d actualRatio=%.5f predictedRatio=%.5f errorBar=%f", pathCtx.String(), d, origLen, finalLen, predictedIntLen, actualRatio, predictedRatio, errorBar)
			}
		}
	}
}

func verifyDistance(pathCtx *PathContext, d int) {
	verifyD := -1
	for _, non := range pathCtx.openEndNodes {
		if !non.pn.IsEnd() {
			verifyD = non.pn.D()
			break
		}
	}
	if verifyD != d {
		Log.Errorf("Something fishy for %s", pathCtx.String())
	}
}
