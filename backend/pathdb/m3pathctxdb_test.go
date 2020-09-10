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
		assert.Equal(t, int64(-3), pn.linkNodeIds[1])
	}

	pn.release()
}

func TestMakeNewPathCtx(t *testing.T) {
	Log.SetAssert(true)
	m3point.Log.SetAssert(true)
	Log.SetDebug()
	m3point.Log.SetDebug()
	m3util.SetToTestMode()
	env := GetFullTestDb(m3util.PathTestEnv)
	start := time.Now()
	InitializePathDBEnv(env)
	endInit := time.Now()
	Log.Infof("Init DB took %v", endInit.Sub(start))

	ppd := pointdb.GetPointPackData(env)

	growthCtx := ppd.GetGrowthContextById(40)
	assert.NotNil(t, growthCtx)
	assert.Equal(t, 40, growthCtx.GetId())
	assert.Equal(t, m3point.GrowthType(8), growthCtx.GetGrowthType())
	assert.Equal(t, 0, growthCtx.GetGrowthIndex())
	pathCtx := MakePathContextDBFromGrowthContext(env, growthCtx, 0)
	assert.NotNil(t, pathCtx)
	assert.Equal(t, 0, pathCtx.GetNumberOfOpenNodes())
	pathCtxDb, ok := pathCtx.(*PathContextDb)
	assert.True(t, ok)
	ctxId := pathCtxDb.id
	assert.True(t, ctxId > 0)
	assert.Nil(t, pathCtxDb.rootNode)

	testPoint := m3point.XFirst
	pathCtx.InitRootNode(testPoint)
	pn := pathCtx.GetRootPathNode()
	assert.NotNil(t, pathCtxDb.rootNode)
	assert.NotNil(t, pn)
	assert.Equal(t, pn, pathCtxDb.rootNode)

	assert.Equal(t, pathCtxDb.rootNode.pathCtxId, ctxId)
	assert.Equal(t, pathCtxDb.rootNode.pointId, getOrCreatePointEnv(env, testPoint))
	assert.Equal(t, 2601, pathCtxDb.rootNode.pathBuilderId)

	assert.Equal(t, 1, pathCtx.GetNumberOfOpenNodes())

	assert.NotNil(t, pathCtxDb.openNodeBuilder)
	assert.Equal(t, pathCtxDb.rootNode, pathCtxDb.openNodeBuilder.openNodesMap.GetPathNode(testPoint))

	assert.Equal(t, 0, pn.D())
	assert.True(t, pn.IsRoot())
	assert.True(t, pn.HasOpenConnections())

	nodeId := pathCtxDb.rootNode.id
	loadedFromDb := pathCtxDb.getPathNodeDb(nodeId)
	assert.NotNil(t, loadedFromDb)
	assert.Equal(t, ctxId, loadedFromDb.pathCtxId)
	assert.Equal(t, pathCtxDb, loadedFromDb.pathCtx)
	assert.Equal(t, nodeId, loadedFromDb.id)
	assert.Equal(t, 2601, loadedFromDb.pathBuilderId)

	Log.Infof("root node is %s", pathCtxDb.rootNode.String())
	Log.Infof("root node from db is %s", loadedFromDb.String())

	rootCreated := time.Now()
	Log.Infof("Total create root DB test took %v", rootCreated.Sub(endInit))

	pathCtx.MoveToNextNodes()
	assert.Equal(t, 3, pathCtx.GetNumberOfOpenNodes())

	moveNext := time.Now()
	Log.Infof("Total move next DB test took %v", moveNext.Sub(rootCreated))

}
