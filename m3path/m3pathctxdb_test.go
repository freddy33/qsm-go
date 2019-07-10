package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMakeNewPathCtx(t *testing.T) {
	Log.SetDebug()
	m3point.Log.SetDebug()
	//m3db.SetToTestMode()
	//env := GetFullTestDb(m3db.PathTestEnv)
	//m3point.SetDefaultEnv(env)
	start := time.Now()
	InitializeDB()
	endInit := time.Now()
	Log.Infof("Init DB took %v", endInit.Sub(start))

	growthCtx := m3point.GetGrowthContextById(40)
	assert.NotNil(t, growthCtx)
	assert.Equal(t, 40, growthCtx.GetId())
	assert.Equal(t, m3point.GrowthType(8), growthCtx.GetGrowthType())
	assert.Equal(t, 0, growthCtx.GetGrowthIndex())
	pathCtx := MakePathContextDBFromGrowthContext(growthCtx, 0, MakeSimplePathNodeMap(12))
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
	assert.Equal(t, pathCtxDb.rootNode.pointId, getOrCreatePoint(testPoint))
	// TODO: Cube index is not always the same
	assert.True(t, 2775 == pathCtxDb.rootNode.pathBuilderId || 2721 == pathCtxDb.rootNode.pathBuilderId, "not right %d", pathCtxDb.rootNode.pathBuilderId)

	assert.Equal(t, 1, pathCtx.GetNumberOfOpenNodes())

	assert.NotNil(t,pathCtxDb.openNodeBuilder)
	assert.Equal(t, pathCtxDb.rootNode, pathCtxDb.openNodeBuilder.openNodes[0])

	assert.Equal(t, 0, pn.D())
	assert.True(t, pn.IsRoot())
	assert.False(t, pn.IsEnd())

	nodeId := pathCtxDb.rootNode.id
	loadedFromDb := getPathNodeDb(nodeId)
	assert.NotNil(t, loadedFromDb)
	assert.Equal(t, ctxId, loadedFromDb.pathCtxId)
	assert.Nil(t, loadedFromDb.pathCtx)
	assert.Equal(t, nodeId, loadedFromDb.id)
	assert.True(t, 2775 == loadedFromDb.pathBuilderId || 2721 == loadedFromDb.pathBuilderId, "not right %d", loadedFromDb.pathBuilderId)

	Log.Infof("root node is %s", pathCtxDb.rootNode.String())
	Log.Infof("root node from db is %s", loadedFromDb.String())

	Log.Infof("Total DB test took %v", time.Now().Sub(endInit))
}