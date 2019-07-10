package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const(
	ExpectedNbConns = 50
	ExpectedNbTrios = 200
	ExpectedNbGrowthContexts = 52
	ExpectedNbCubes = 5192
	ExpectedNbPathBuilders = ExpectedNbCubes
)

func TestLoadOrCalculate(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()
	m3db.SetToTestMode()

	env := GetFullTestDb(m3db.PointTestEnv)

	start := time.Now()
	allConnections, allConnectionsByVector = calculateConnectionDetails()
	connectionsLoaded[pointEnvId] = true
	allTrioDetails = calculateAllTrioDetails()
	trioDetailsLoaded[pointEnvId] = true
	allGrowthContexts = calculateAllGrowthContexts()
	growthContextsLoaded[pointEnvId] = true
	cubeIdsPerKey = calculateAllContextCubes()
	cubesLoaded[pointEnvId] = true
	pathBuilders = calculateAllPathBuilders()
	pathBuildersLoaded[pointEnvId] = true
	calcTime := time.Now().Sub(start)
	Log.Infof("Took %v to calculate", calcTime)

	assert.Equal(t, ExpectedNbConns, len(allConnections))
	assert.Equal(t, ExpectedNbConns, len(allConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(allTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(allGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(cubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(pathBuilders)-1)

	start = time.Now()
	// force reload
	InitializeDBEnv(env, true)
	loadTime := time.Now().Sub(start)
	Log.Infof("Took %v to load", loadTime)

	Log.Infof("Diff calc-load = %v", calcTime - loadTime)

	assert.Equal(t, ExpectedNbConns, len(allConnections))
	assert.Equal(t, ExpectedNbConns, len(allConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(allTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(allGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(cubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(pathBuilders)-1)
}

func TestSaveAll(t *testing.T) {
	m3db.Log.SetDebug()
	Log.SetDebug()
	m3db.SetToTestMode()

	tempEnv := GetCleanTempDb(m3db.PointTempEnv)

	// ************ Connection Details

	n, err := saveAllConnectionDetails(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbConns, n)

	// Should be able to run twice
	n, err = saveAllConnectionDetails(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbConns, n)

	// Test we can load
	loaded, _ := loadConnectionDetails(tempEnv)
	assert.Equal(t, ExpectedNbConns, len(loaded))

	// Init
	initConnections(tempEnv)

	// ************ Trio Details

	n, err = saveAllTrioDetails(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Should be able to run twice
	n, err = saveAllTrioDetails(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Test we can load
	loaded2 := loadTrioDetails(tempEnv)
	assert.Equal(t, ExpectedNbTrios, len(loaded2))

	// Init
	initTrioDetails(tempEnv)

	// ************ Growth Contexts

	n, err = saveAllGrowthContexts(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Should be able to run twice
	n, err = saveAllGrowthContexts(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Test we can load
	loaded3 := loadGrowthContexts(tempEnv)
	assert.Equal(t, ExpectedNbGrowthContexts, len(loaded3))

	// Init
	initGrowthContexts(tempEnv)

	// ************ Context Cubes

	n, err = saveAllContextCubes(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Should be able to run twice
	n, err = saveAllContextCubes(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Test we can load
	loaded4 := loadContextCubes(tempEnv)
	assert.Equal(t, ExpectedNbCubes, len(loaded4))

	// Init
	initContextCubes(tempEnv)

	// ************ Path Builders

	n, err = saveAllPathBuilders(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Should be able to run twice
	n, err = saveAllPathBuilders(tempEnv)
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Test we can load
	loaded5 := loadPathBuilders(tempEnv)
	assert.Equal(t, ExpectedNbPathBuilders, len(loaded5)-1)

	// Init from Good DB
	initPathBuilders(tempEnv)
}
