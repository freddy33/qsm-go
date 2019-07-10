package m3path

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeNewPathCtx(t *testing.T) {
	m3db.SetToTestMode()
	env := GetFullTestDb(m3db.PathTestEnv)
	m3point.SetDefaultEnv(env)
	growthCtx := m3point.GetGrowthContextById(0)
	assert.NotNil(t, growthCtx)
	assert.Equal(t, 0, growthCtx.GetId())
	assert.Equal(t, m3point.GrowthType(1), growthCtx.GetGrowthType())
	assert.Equal(t, 0, growthCtx.GetGrowthIndex())
	pathCtx := MakePathContextDBFromGrowthContext(growthCtx, 0, MakeSimplePathNodeMap(12))
	assert.NotNil(t, pathCtx)
	pathCtxDb, ok := pathCtx.(*PathContextDb)
	assert.True(t, ok)
	assert.True(t, pathCtxDb.id > 0)
	assert.Nil(t, pathCtxDb.rootNode)
	testPoint := m3point.XFirst
	pathCtx.InitRootNode(testPoint)
	assert.NotNil(t, pathCtxDb.rootNode)
	assert.Equal(t, pathCtxDb.rootNode.pathCtxId, pathCtxDb.id)
	assert.Equal(t, pathCtxDb.rootNode.pointId, getOrCreatePoint(testPoint))
}