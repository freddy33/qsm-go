package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

var allTestContexts map[m3point.ContextType][]*PathContext

func getAllTestContexts() map[m3point.ContextType][]*PathContext {
	if allTestContexts != nil {
		return allTestContexts
	}
	res := make(map[m3point.ContextType][]*PathContext)

	m3point.InitializeDetails()

	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		maxOffset := m3point.MaxOffsetPerType[ctxType]
		res[ctxType] = make([]*PathContext, nbIndexes*maxOffset)
		idx := 0
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			for offset := 0; offset < maxOffset; offset++ {
				res[ctxType][idx] = MakePathContext(ctxType, pIdx, offset, nil)
				idx++
			}
		}
	}

	allTestContexts = res
	return res
}

func TestFirstPathContextFilling(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetInfo()
	m3point.Log.SetAssert(true)

	allCtx := getAllTestContexts()
	for _, ctxType := range m3point.GetAllContextTypes() {
		for _, ctx := range allCtx[ctxType] {
			pathCtx := MakePathContext(ctxType, ctx.GetIndex(), ctx.offset, MakeSimplePathNodeMap(2^4))
			fillPathContext(t, pathCtx, 8*3)
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

	pathNodeMap := pathCtx.GetPathNodeMap()
	assert.Equal(t, 1+3, pathNodeMap.GetSize(), "not all points are here %v", pathCtx.openEndNodes)
	assert.Equal(t, 3, len(pathCtx.openEndNodes), "not all ends here %v", pathCtx.openEndNodes)
	spnm, ok := pathNodeMap.(*SimplePathNodeMap)
	assert.True(t, ok, "should be a simple path node map for %v", pathNodeMap)
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
		assert.True(t, oep.pn.IsLatest(), "open end path %v should be active", oep.pn)
	}
	assert.Equal(t, 0, countMains, "not all main ends here %v", pathCtx.openEndNodes)
	assert.Equal(t, 3, countNonMains, "not all non main ends here %v", pathCtx.openEndNodes)

	if until == 2 {
		Log.Debug("*************** First round *************\n", pathCtx.dumpInfo())
		pathCtx.MoveToNextNodes()
		assertPathContextState(t, pathCtx, spnm)
		Log.Debug("*************** Second round *************\n", pathCtx.dumpInfo())
	} else {
		for d := 1; d < until; d++ {
			pathCtx.MoveToNextNodes()
			assertPathContextState(t, pathCtx, spnm)
		}
	}
}

func assertPathContextState(t *testing.T, pathCtx *PathContext, spnm *SimplePathNodeMap) {
	inOpenEnd := make(map[m3point.Point]bool)
	for _, oep := range pathCtx.openEndNodes {
		assert.NotEqual(t, m3point.NilTrioIndex, oep.pn.GetTrioIndex(), "%v should have trio already", oep.pn)
		assert.Equal(t, oep.pn.calcDist(), oep.pn.D(), "open end path %v should have d and calcDist equal", oep.pn)
		// TODO: Find a way to test that open end node are mostly active
		//assert.True(t, oep.pn.IsLatest(), "open end path %v should be active", oep.pn)
		inOpenEnd[oep.pn.P()] = true
	}
	for p, n := range *spnm {
		if !inOpenEnd[p] {
			assert.False(t, n.IsLatest(), "non open end path %v should be active", n)
		}
	}
}
