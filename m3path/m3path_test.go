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
			pathCtx := MakePathContext(ctxType, ctx.GetIndex(), ctx.offset, MakeSimplePathNodeMap(2^4))
			fillPathContext(t, pathCtx, 4)
			Log.Infof("Run for %s got %d points %d last open end path", pathCtx.String(), pathCtx.GetPathNodeMap().GetSize(), len(pathCtx.openEndNodes))
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

	pathCtx.InitRootNode(m3point.Origin)
	pathCtx.MoveToNextNodes()

	assert.Equal(t, 1+3, pathCtx.GetPathNodeMap().GetSize(), "not all points are here %v", pathCtx.openEndNodes)
	assert.Equal(t, 3, len(pathCtx.openEndNodes), "not all ends here %v", pathCtx.openEndNodes)
	countMains := 0
	countNonMains := 0
	for _, oep := range pathCtx.openEndNodes {
		assert.NotEqual(t, m3point.NilTrioIndex, oep.pn.GetTrioIndex(), "%v should have trio already", oep.pn)
		if oep.pn.P().IsMainPoint() {
			countMains++
		} else {
			countNonMains++
		}
		assert.Equal(t, 1, oep.pn.D(), "open end path %v should have distance of three", oep.pn)
		assert.Equal(t, oep.pn.calcDist(), oep.pn.D(), "open end path %v should have d and calcDist equal", oep.pn)
	}
	assert.Equal(t, 0, countMains, "not all main ends here %v", pathCtx.openEndNodes)
	assert.Equal(t, 3, countNonMains, "not all non main ends here %v", pathCtx.openEndNodes)

	if until == 2 {
		Log.Debug("*************** First round *************\n", pathCtx.dumpInfo())
		pathCtx.MoveToNextNodes()
		Log.Debug("*************** Second round *************\n", pathCtx.dumpInfo())
	} else {
		for d := 1; d < until; d++ {
			pathCtx.MoveToNextNodes()
		}
	}
}
