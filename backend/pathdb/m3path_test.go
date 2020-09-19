package pathdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var allTestContextsMutex sync.Mutex

func getAllTestContexts(env m3util.QsmEnvironment) map[m3point.GrowthType][]m3path.PathContext {
	pathData := GetServerPathPackData(env).(*ServerPathPackData)
	if pathData.AllCenterContextsLoaded {
		return pathData.AllCenterContexts
	}

	allTestContextsMutex.Lock()
	defer allTestContextsMutex.Unlock()

	if pathData.AllCenterContextsLoaded {
		return pathData.AllCenterContexts
	}

	pointdb.InitializePointDBEnv(env.(*m3db.QsmDbEnvironment), false)
	pointData := pointdb.GetPointPackData(env)

	idx := 0
	for _, growthCtx := range pointData.GetAllGrowthContexts() {
		ctxType := growthCtx.GetGrowthType()
		maxOffset := ctxType.GetMaxOffset()
		if len(pathData.AllCenterContexts[ctxType]) == 0 {
			pathData.AllCenterContexts[ctxType] = make([]m3path.PathContext, ctxType.GetNbIndexes()*maxOffset)
			idx = 0
		}
		for offset := 0; offset < maxOffset; offset++ {
			pathData.AllCenterContexts[ctxType][idx] = MakePathContextDBFromGrowthContext(env, growthCtx, offset)
			idx++
		}
	}

	pathData.AllCenterContextsLoaded = true
	return pathData.AllCenterContexts
}

func TestFirstPathContextFilling(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetInfo()
	m3point.Log.SetAssert(true)
	m3util.SetToTestMode()

	env := GetFullTestDb(m3util.PathTestEnv)
	allCtx := getAllTestContexts(env)
	for _, ctxType := range m3point.GetAllGrowthTypes() {
		for _, ctx := range allCtx[ctxType] {
			fillPathContext(t, ctx, 12)
			Log.Infof("Run for %s got %d points %d last open end path", ctx.String(), ctx.CountAllPathNodes(), ctx.GetNumberOfOpenNodes())
			if Log.IsDebug() {
				Log.Debug(ctx.DumpInfo())
			}
			break
		}
	}
}

func fillPathContext(t *testing.T, pathCtx m3path.PathContext, until int) {
	growthCtx := pathCtx.GetGrowthCtx()
	ppd := pointdb.GetPointPackData(growthCtx.GetEnv())
	trIdx := growthCtx.GetBaseTrioIndex(ppd, 0, pathCtx.GetGrowthOffset())
	assert.NotEqual(t, m3point.NilTrioIndex, trIdx)

	td := ppd.GetTrioDetails(trIdx)
	assert.NotNil(t, td)
	assert.Equal(t, trIdx, td.GetId())

	Log.Debug(growthCtx.String(), td.String())

	pathCtx.InitRootNode(m3point.Origin)
	pathCtx.MoveToNextNodes()

	//pathNodeMap := pathCtx.GetPathNodeMap()
	assert.Equal(t, 1+3, pathCtx.CountAllPathNodes(), "not all points of %s are here", pathCtx.String())
	assert.Equal(t, 3, pathCtx.GetNumberOfOpenNodes(), "not all ends of %s here", pathCtx.String())
	//spnm, ok := pathNodeMap.(*SimplePathNodeMap)
	//assert.True(t, ok, "should be a simple path node map for %v", pathNodeMap)
	countMains := 0
	countNonMains := 0
	openEndNodes := pathCtx.GetAllOpenPathNodes()
	for _, pn := range openEndNodes {
		assert.NotEqual(t, m3point.NilTrioIndex, pn.GetTrioIndex(), "%v should have trio already", pn)
		if pn.P().IsMainPoint() {
			countMains++
		} else {
			countNonMains++
		}
		assert.Equal(t, 1, pn.D(), "open end path %v should have distance of three", pn)
		assert.True(t, pn.IsLatest(), "open end path %v should be active", pn)
	}
	assert.Equal(t, 0, countMains, "not all main ends here %v", openEndNodes)
	assert.Equal(t, 3, countNonMains, "not all non main ends here %v", openEndNodes)

	if until == 2 {
		Log.Debug("*************** First round *************\n", pathCtx.DumpInfo())
		pathCtx.MoveToNextNodes()
		assertPathContextState(t, pathCtx.GetAllOpenPathNodes())
		Log.Debug("*************** Second round *************\n", pathCtx.DumpInfo())
	} else {
		for d := 1; d < until; d++ {
			pathCtx.MoveToNextNodes()
			assertPathContextState(t, pathCtx.GetAllOpenPathNodes())
		}
	}
}

func assertPathContextState(t *testing.T, openEndNodes []m3path.PathNode) {
	//inOpenEnd := make(map[m3point.Point]bool)
	for _, pn := range openEndNodes {
		assert.True(t, pn.(*PathNodeDb).id > 0, "%v should have an id already", pn)
		assert.NotEqual(t, m3point.NilTrioIndex, pn.GetTrioIndex(), "%v should have trio already", pn)
		//assert.Equal(t, pn.calcDist(), pn.D(), "open end path %v should have d and calcDist equal", pn)
		// TODO: Find a way to test that open end node are mostly active
		//assert.True(t, oep.pn.IsLatest(), "open end path %v should be active", oep.pn)
		//inOpenEnd[pn.P()] = true
	}
	//for p, n := range *spnm {
	//	if !inOpenEnd[p] {
	//		assert.False(t, n.IsLatest(), "non open end path %v should be active", n)
	//	}
	//}
}
