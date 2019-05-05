package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
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
// PathContext Bench
/***************************************************************/

func BenchmarkPathCtx3(b *testing.B) {
	runForPathCtxType(b.N, BenchNbRound, 3)
}

func BenchmarkPathCtx4(b *testing.B) {
	runForPathCtxType(b.N, BenchNbRound, 4)
}

func BenchmarkPathCtx8(b *testing.B) {
	runForPathCtxType(b.N, BenchNbRound, 8)
}

func runForPathCtxType(N, until int, pType m3point.ContextType) {
	Log.SetWarn()
	Log.SetAssert(true)
	m3point.Log.SetWarn()
	m3point.Log.SetAssert(true)

	allCtx := getAllTestContexts()
	for r := 0; r < N; r++ {
		for _, ctx := range allCtx[pType] {
			start := time.Now()
			pathCtx := MakePathContext(ctx.GetType(), ctx.GetIndex(), ctx.offset)
			pathCtx.pathNodesPerPoint = make(map[m3point.Point]*PathNode, 5*until*until)
			runPathContext(pathCtx, until/3)
			t := time.Since(start)
			LogDataTest.Infof("%s %s %d %d %d", t, pathCtx, len(pathCtx.pathNodesPerPoint), len(pathCtx.openEndPaths), pathCtx.openEndPaths[0].pn.d)
		}
	}
}

func runPathContext(pathCtx *PathContext, until int) {
	pathCtx.initRootLinks()
	for d := 0; d < until; d++ {
		pathCtx.moveToNextMainPoints()
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
