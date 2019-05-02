package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"testing"
)

func BenchmarkPathCtx3(b *testing.B) {
	runForPathCtxType(b.N, BENCH_NB_ROUND, 3)
}

func BenchmarkPathCtx4(b *testing.B) {
	runForPathCtxType(b.N, BENCH_NB_ROUND, 4)
}

func BenchmarkPathCtx8(b *testing.B) {
	runForPathCtxType(b.N, BENCH_NB_ROUND, 8)
}

func runForPathCtxType(N, until int, pType m3point.ContextType) {
	Log.SetWarn()
	Log.SetAssert(true)
	m3point.Log.SetWarn()
	m3point.Log.SetAssert(true)

	allCtx := getAllTestContexts()
	for r := 0; r < N; r++ {
		for _, ctx := range allCtx[pType] {
			pathCtx := MakePathContext(ctx.GetType(), ctx.GetIndex(), ctx.offset)
			pathCtx.pathNodesPerPoint = make(map[m3point.Point]*PathNode, 5*until*until)
			runPathContext(pathCtx, until/3)
		}
	}
}

func runPathContext(pathCtx *PathContext, until int) {
	pathCtx.initRootLinks()
	for d := 0; d < until; d++ {
		pathCtx.moveToNextMainPoints()
	}
	Log.Infof("Run for %s got %d points %d last open end path", pathCtx.String(), len(pathCtx.pathNodesPerPoint), len(pathCtx.openEndPaths))
}

const (
	SPLIT          = 4
	BENCH_NB_ROUND = 90
	TEST_NB_ROUND  = 25
)

/*
func BenchmarkGrowthCtx1(b *testing.B) {
	Log.SetWarn()
	runForCtxType(b.N, BENCH_NB_ROUND, 1)
}

func BenchmarkGrowthCtx2(b *testing.B) {
	Log.SetWarn()
	runForCtxType(b.N, BENCH_NB_ROUND, 2)
}
*/

func BenchmarkGrowthCtx3(b *testing.B) {
	Log.SetWarn()
	runForCtxType(b.N, BENCH_NB_ROUND, 3)
}

func BenchmarkGrowthCtx4(b *testing.B) {
	Log.SetWarn()
	runForCtxType(b.N, BENCH_NB_ROUND, 4)
}

func BenchmarkGrowthCtx8(b *testing.B) {
	Log.SetWarn()
	runForCtxType(b.N, BENCH_NB_ROUND, 8)
}

