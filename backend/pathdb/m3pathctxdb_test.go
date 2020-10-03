package pathdb

import (
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNodeSyncPool(t *testing.T) {
	pn := getNewPathNodeDb()
	assert.NotNil(t, pn)
	assert.Equal(t, int64(-1), pn.id)
	assert.Equal(t, -1, pn.pathCtxId)
	for i := 0; i < 3; i++ {
		assert.Equal(t, int64(-3), pn.LinkIds[1])
	}

	pn.release()
}

func TestMakeNewPathCtx(t *testing.T) {
	Log.SetAssert(true)
	m3point.Log.SetAssert(true)
	Log.SetDebug()
	m3point.Log.SetDebug()
	m3util.SetToTestMode()
	env := GetPathDbFullEnv(m3util.PathTestEnv)

	pointData := pointdb.GetServerPointPackData(env)
	pathData := GetServerPathPackData(env)

	growthCtx := pointData.GetGrowthContextById(40)
	assert.NotNil(t, growthCtx)
	assert.Equal(t, 40, growthCtx.GetId())
	assert.Equal(t, m3point.GrowthType(8), growthCtx.GetGrowthType())
	assert.Equal(t, 0, growthCtx.GetGrowthIndex())
	pathCtx, err := pathData.GetPathCtxDbFromAttributes(growthCtx.GetGrowthType(), growthCtx.GetGrowthIndex(), 0)
	assert.NoError(t, err)
	assert.NotNil(t, pathCtx)
	assert.Equal(t, 1, pathCtx.GetNumberOfNodesAt(0))
	ctxId := pathCtx.id
	assert.True(t, ctxId > 0)

	testPoint := m3point.Origin
	pn := pathCtx.GetRootPathNode()
	assert.NotNil(t, pathCtx.rootNode)
	assert.NotNil(t, pn)
	assert.Equal(t, pn, pathCtx.rootNode)

	assert.Equal(t, pathCtx.rootNode.pathCtxId, ctxId)
	assert.Equal(t, pathCtx.rootNode.pointId, pathData.GetOrCreatePoint(testPoint))
	assert.Equal(t, 2601, pathCtx.rootNode.pathBuilderId)

	assert.Equal(t, 1, pathCtx.GetNumberOfNodesAt(0))

	assert.Equal(t, 0, pn.D())
	assert.True(t, pn.IsRoot())

	nodeId := pathCtx.rootNode.id
	loadedFromDb, err := pathCtx.GetPathNodeDb(nodeId)
	assert.NoError(t, err)
	assert.NotNil(t, loadedFromDb)
	assert.Equal(t, ctxId, loadedFromDb.pathCtxId)
	assert.Equal(t, pathCtx, loadedFromDb.pathCtx)
	assert.Equal(t, nodeId, loadedFromDb.id)
	assert.Equal(t, 2601, loadedFromDb.pathBuilderId)

	Log.Infof("root node is %s", pathCtx.rootNode.String())
	Log.Infof("root node from db is %s", loadedFromDb.String())

	rootCreated := time.Now()

	err = pathCtx.calculateNextMaxDist()
	assert.NoError(t, err)
	assert.Equal(t, 3, pathCtx.GetNumberOfNodesAt(1))

	moveNext := time.Now()
	Log.Infof("Total move next DB test took %v", moveNext.Sub(rootCreated))

}
