package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFirstPathContextFilling(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetInfo()
	m3point.Log.SetAssert(true)

	allCtx := getAllTestContexts()
	for _, ctxType := range m3point.GetAllContextTypes() {
		for _, ctx := range allCtx[ctxType] {
			pathCtx := MakePathContext(ctxType, ctx.GetIndex(), ctx.offset)
			fillPathContext(t, pathCtx, 1)
			Log.Infof("Run for %s got %d points %d last open end path", pathCtx.String(), len(pathCtx.pathNodesPerPoint), len(pathCtx.openEndPaths))
			Log.Debug( pathCtx.dumpInfo())
			break
		}
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
