package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
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

func TestFirstPathContextFilling(t *testing.T) {
	Log.SetTrace()
	Log.SetAssert(true)
	m3point.Log.SetTrace()
	m3point.Log.SetAssert(true)
	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			pathCtx := MakePathContext(ctxType, pIdx, 0)
			fillPathContext(t, pathCtx, 2)
			break
		}
	}
}

func TestAllPathContextFilling(t *testing.T) {
	Log.SetWarn()
	Log.SetAssert(true)
	m3point.Log.SetWarn()
	m3point.Log.SetAssert(true)
	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			pathCtx := MakePathContext(ctxType, pIdx, 0)
			fillPathContext(t, pathCtx, 8*2)
			Log.Warnf("Run for %s got %d points %d last open end path", pathCtx.String(), len(pathCtx.pathNodesPerPoint), len(pathCtx.openEndPaths))
		}
	}
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
			Log.Infof("Run for %s got %d points %d last open end path", pathCtx.String(), len(pathCtx.pathNodesPerPoint), len(pathCtx.openEndPaths))
		}
	}
}

func runPathContext(pathCtx *PathContext, until int) {
	pathCtx.initRootLinks()
	for d := 0; d < until; d++ {
		pathCtx.moveToNextMainPoints()
	}
}

func fillPathContext(t *testing.T, pathCtx *PathContext, until int) {
	trCtx := pathCtx.ctx
	trIdx := trCtx.GetBaseTrioIndex(0, pathCtx.offset)
	assert.NotEqual(t, m3point.NilTrioIndex, trIdx)

	td := m3point.GetTrioDetails(trIdx)
	assert.NotNil(t, td)
	assert.Equal(t, trIdx, td.GetId())

	Log.Debug(trCtx.String(), td.String())

	pathCtx.initRootLinks()
	pathCtx.moveToNextMainPoints()

	assert.Equal(t, 1+3+6+12, len(pathCtx.pathNodesPerPoint), "not all points are here %v", pathCtx.openEndPaths)
	assert.Equal(t, 12, len(pathCtx.openEndPaths), "not all ends here %v", pathCtx.openEndPaths)
	countMains := 0
	countNonMains := 0
	for _, oep := range pathCtx.openEndPaths {
		assert.Equal(t, oep.kind == MainPointOpenPath, oep.pn.p.IsMainPoint(), "main bool for %v should be equal to point is main()", *oep.pn)
		if oep.kind == MainPointOpenPath {
			countMains++
			assert.NotEqual(t, m3point.NilTrioIndex, oep.pn.trioId, "main %v should have trio already", *oep.pn)
		} else {
			countNonMains++
			assert.Equal(t, m3point.NilTrioIndex, oep.pn.trioId, "non main %v should not have trio already", *oep.pn)
		}
		assert.Equal(t, 3, oep.pn.d, "open end path %v should have distance of three", *oep.pn)
		assert.Equal(t, oep.pn.calcDist(), oep.pn.d, "open end path %v should have d and calcDist equal", *oep.pn)
	}
	assert.Equal(t, 6, countMains, "not all main ends here %v", pathCtx.openEndPaths)
	assert.Equal(t, 6, countNonMains, "not all non main ends here %v", pathCtx.openEndPaths)

	if until == 2 {
		Log.Debug("*************** First round *************\n", pathCtx.dumpInfo())
		pathCtx.moveToNextMainPoints()
		Log.Debug("*************** Second round *************\n", pathCtx.dumpInfo())
	} else {
		for d := 1; d < until; d++ {
			pathCtx.moveToNextMainPoints()
		}
	}
}
