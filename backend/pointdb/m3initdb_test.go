package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
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
	m3util.SetToTestMode()

	env := GetPointDbFullEnv(m3util.PointLoadEnv)
	ppd, _ := GetServerPointPackData(env)

	start := time.Now()
	ppd.ResetFlags()
	ppd.AllConnections, ppd.AllConnectionsByVector = ppd.calculateConnectionDetails()
	ppd.ConnectionsLoaded = true
	ppd.AllTrioDetails = ppd.calculateAllTrioDetails()
	ppd.TrioDetailsLoaded = true
	ppd.AllGrowthContexts = ppd.calculateAllGrowthContexts()
	ppd.GrowthContextsLoaded = true
	ppd.cubeIdsPerKey = ppd.calculateAllContextCubes()
	ppd.cubesLoaded = true
	ppd.pathBuilders = ppd.calculateAllPathBuilders()
	ppd.pathBuildersLoaded = true
	calcTime := time.Now().Sub(start)
	Log.Infof("Took %v to calculate", calcTime)

	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(ppd.AllTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.AllGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(ppd.cubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.pathBuilders)-1)

	start = time.Now()
	ppd.AllConnections = nil
	ppd.AllConnectionsByVector = nil
	ppd.AllTrioDetails = nil
	ppd.AllGrowthContexts = nil
	ppd.cubeIdsPerKey = nil
	ppd.pathBuilders = nil
	ppd.ResetFlags()
	ppd.FillDb()
	fillDbTime := time.Now().Sub(start)
	Log.Infof("Took %v to fill db", fillDbTime)

	start = time.Now()
	// force reload
	ppd.ResetFlags()
	ppd.InitializeAll()
	loadTime := time.Now().Sub(start)
	Log.Infof("Took %v to load", loadTime)

	Log.Infof("Diff calc-load = %v", calcTime-loadTime)

	// Don't forget to get ppd different after init
	ppd, _ = GetServerPointPackData(env)
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(ppd.AllTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.AllGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(ppd.cubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.pathBuilders)-1)
}

func TestSaveAll(t *testing.T) {
	m3db.Log.SetDebug()
	Log.SetDebug()
	m3util.SetToTestMode()

	tempEnv := GetPointDbCleanEnv(m3util.PointTempEnv)
	ppd, _ := GetServerPointPackData(tempEnv)

	// ************ Connection Details

	n, err := ppd.saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbConns, n)

	// Should be able to run twice
	n, err = ppd.saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbConns, n)

	// Test we can load
	err = ppd.loadConnectionDetails()
	assert.NoError(t, err)
	assert.True(t, ppd.ConnectionsLoaded)
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnectionsByVector))

	// Init twice
	ppd.ConnectionsLoaded = false
	ppd.initConnections()

	// ************ Trio Details

	n, err = ppd.saveAllTrioDetails()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Should be able to run twice
	n, err = ppd.saveAllTrioDetails()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Test we can load
	err = ppd.loadTrioDetails()
	assert.NoError(t, err)
	assert.True(t, ppd.TrioDetailsLoaded)
	assert.Equal(t, ExpectedNbTrios, len(ppd.AllTrioDetails))

	// Init twice
	ppd.TrioDetailsLoaded = false
	ppd.initTrioDetails()

	// ************ Growth Contexts

	n, err = ppd.saveAllGrowthContexts()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Should be able to run twice
	n, err = ppd.saveAllGrowthContexts()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Test we can load
	err = ppd.loadGrowthContexts()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.AllGrowthContexts))
	assert.True(t, ppd.GrowthContextsLoaded)

	// Init twice
	ppd.GrowthContextsLoaded = false
	ppd.initGrowthContexts()

	// ************ Context Cubes

	n, err = ppd.saveAllContextCubes()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Should be able to run twice
	n, err = ppd.saveAllContextCubes()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Test we can load
	err = ppd.loadContextCubes()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbCubes, len(ppd.cubeIdsPerKey))
	assert.True(t, ppd.cubesLoaded)

	// Init twice
	ppd.cubesLoaded = false
	ppd.initContextCubes()

	// ************ Path Builders

	n, err = ppd.saveAllPathBuilders()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Should be able to run twice
	n, err = ppd.saveAllPathBuilders()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Test we can load
	err = ppd.loadPathBuilders()
	assert.NoError(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.pathBuilders)-1)
	assert.True(t, ppd.pathBuildersLoaded)

	// Init from Good DB
	ppd.pathBuildersLoaded = false
	ppd.initPathBuilders()
}
