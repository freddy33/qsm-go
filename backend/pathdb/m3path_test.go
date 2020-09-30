package pathdb

import (
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFirstPathContextFilling(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetInfo()
	m3point.Log.SetAssert(true)
	m3util.SetToTestMode()

	env := GetPathDbFullEnv(m3util.PathTestEnv)
	pathData := GetServerPathPackData(env)
	for _, ctxType := range m3point.GetAllGrowthTypes() {
		for _, pathCtx := range pathData.AllCenterContexts[ctxType] {
			until := 12
			fillPathContext(t, pathCtx, until)
			Log.Infof("Run for %s got %d points %d last open end path", pathCtx.String(), pathCtx.CountAllPathNodes(), pathCtx.GetNumberOfNodesAt(until))
			if Log.IsDebug() {
				Log.Debug(pathCtx.DumpInfo())
			}
			break
		}
	}
}

func fillPathContext(t *testing.T, pathCtx *PathContextDb, until int) {
	growthCtx := pathCtx.GetGrowthCtx()
	ppd := pointdb.GetPointPackData(growthCtx.GetEnv())
	trIdx := growthCtx.GetBaseTrioIndex(ppd, 0, pathCtx.GetGrowthOffset())
	assert.NotEqual(t, m3point.NilTrioIndex, trIdx)

	td := ppd.GetTrioDetails(trIdx)
	assert.NotNil(t, td)
	assert.Equal(t, trIdx, td.GetId())

	Log.Debug(growthCtx.String(), td.String())

	for d := 0; d <= until; d++ {
		if d > pathCtx.GetMaxDist() {
			err := pathCtx.calculateNextMaxDist()
			assert.NoError(t, err)
		}
	}
}
