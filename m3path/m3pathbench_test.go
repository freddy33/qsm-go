package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"math"
	"sync"
	"testing"
	"time"
)

var LogDataTest = m3util.NewDataLogger("DATA", m3util.DEBUG)

const (
	BenchNbRound    = 51
	MinSizePerSplit = 16
)

/***************************************************************/
// Compare PathContext to Growth Context
/***************************************************************/

func TestComparePathToGrowth(t *testing.T) {
	contextType := m3point.ContextType(4)
	contextIndex := 0
	contextOffset := 0
	nbRound := 8*3

	// Create the 2 kind of context identical
	trioCtx := m3point.GetTrioIndexContext(contextType, contextIndex)
	growthCtx := CreateFromRoot(trioCtx, m3point.Origin, contextOffset)
	pathCtx := MakePathContextFromTrioContext(trioCtx, contextOffset, MakeSimplePathNodeMap(5*nbRound*nbRound))

	usedPoints := make(map[m3point.Point]int, 5*nbRound*nbRound)

	// Initialize the first entries
	latestPoints := make([]m3point.Point, 1)
	latestPoints[0] = m3point.Origin
	usedPoints[m3point.Origin] = 0

	pathCtx.InitRootNode(m3point.Origin)

	for d := 1; d < nbRound; d++ {
		pnm := *(pathCtx.GetPathNodeMap().(*SimplePathNodeMap))
		for p, pp := range pnm {
			opd, ok := usedPoints[p]
			assert.True(t, ok, "did not find point %v at %d in used points %v", p, d, usedPoints)
			assert.Equal(t, opd, pp.D(), "dist of point %v not equal at %d", p, d)
		}
		pathCtx.MoveToNextNodes()

		nbLatestPoints := len(latestPoints)
		finalPoints := make([]m3point.Point, 0, int(1.7*float32(nbLatestPoints)))
		for _, lp := range latestPoints {
			for _, np := range growthCtx.GetNextPoints(lp) {
				existingD, ok := usedPoints[np]
				if !ok {
					finalPoints = append(finalPoints, np)
					usedPoints[np] = d
				} else {
					if d - existingD > 2 {
						Log.Infof("New point %v on existing one with not same dist %d != %d", np, existingD, d)
					}
				}
			}
		}
		latestPoints = finalPoints
	}
}

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

/***************************************************************/
// GrowthContext Bench
/***************************************************************/

func BenchmarkGrowthCtx3Split4(b *testing.B) {
	Log.SetWarn()
	runForCtxTypeSplit(b.N, BenchNbRound, 3, 4)
}

func BenchmarkGrowthCtx4Split4(b *testing.B) {
	Log.SetWarn()
	runForCtxTypeSplit(b.N, BenchNbRound, 4, 4)
}

func BenchmarkGrowthCtx8Split4(b *testing.B) {
	Log.SetWarn()
	runForCtxTypeSplit(b.N, BenchNbRound, 8, 4)
}

func runForCtxTypeSplit(N, nbRound int, pType m3point.ContextType, split int) {
	allCtx := getAllTestContexts()
	for r := 0; r < N; r++ {
		for _, ctx := range allCtx[pType] {
			start := time.Now()
			nU, nL := runNextPoints(&ctx, nbRound)
			t := time.Since(start)
			LogDataTest.Infof("%s %s %d %d %d", t, ctx.String(), nU, nL, nbRound)
		}
	}
}

func runNextPointsSplit(ctx *GrowthContext, nbRound int, split int) (int, int) {
	usedPoints := make(map[m3point.Point]bool, 5*nbRound*nbRound)
	totalUsedPoints := 1
	latestPoints := make([]m3point.Point, 1)
	latestPoints[0] = m3point.Origin
	usedPoints[m3point.Origin] = true
	for d := 0; d < nbRound; d++ {
		nbLatestPoints := len(latestPoints)
		// Send all orig new points
		origNewPoints := make(chan m3point.Point, MinSizePerSplit*split)
		wg := sync.WaitGroup{}
		if nbLatestPoints < MinSizePerSplit*split {
			// too small for split send all
			wg.Add(1)
			go nextPointsSplit(&latestPoints, 0, nbLatestPoints, ctx, origNewPoints, &wg)
		} else {
			sizePerSplit := int(nbLatestPoints / split)
			for currentPos := 0; currentPos < nbLatestPoints; currentPos += sizePerSplit {
				wg.Add(1)
				go nextPointsSplit(&latestPoints, currentPos, sizePerSplit, ctx, origNewPoints, &wg)
			}
		}
		go func(step int) {
			wg.Wait()
			close(origNewPoints)
		}(d)

		finalPoints := make([]m3point.Point, 0, int(1.7*float32(nbLatestPoints)))
		for p := range origNewPoints {
			_, ok := usedPoints[p]
			if !ok {
				finalPoints = append(finalPoints, p)
				usedPoints[p] = true
			}
		}

		totalUsedPoints += len(finalPoints)
		latestPoints = finalPoints
	}
	return totalUsedPoints, len(latestPoints)
}

func nextPointsSplit(lps *[]m3point.Point, currentPos, nb int, ctx *GrowthContext, o chan m3point.Point, wg *sync.WaitGroup) {
	c := 0
	for i := currentPos; i < len(*lps); i++ {
		p := (*lps)[i]
		for _, np := range ctx.GetNextPoints(p) {
			o <- np
		}
		c++
		if c == nb {
			break
		}
	}
	wg.Done()
}
