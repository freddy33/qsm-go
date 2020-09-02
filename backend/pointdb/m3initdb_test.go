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

	env := GetServerFullTestDb(m3util.PointLoadEnv)
	ppd, _ := GetServerPointPackData(env)

	start := time.Now()
	ppd.ResetFlags()
	ppd.AllConnections, ppd.AllConnectionsByVector = ppd.calculateConnectionDetails()
	ppd.ConnectionsLoaded = true
	ppd.AllTrioDetails = ppd.calculateAllTrioDetails()
	ppd.TrioDetailsLoaded = true
	ppd.AllGrowthContexts = ppd.calculateAllGrowthContexts()
	ppd.GrowthContextsLoaded = true
	ppd.CubeIdsPerKey = ppd.calculateAllContextCubes()
	ppd.CubesLoaded = true
	ppd.PathBuilders = ppd.calculateAllPathBuilders()
	ppd.PathBuildersLoaded = true
	calcTime := time.Now().Sub(start)
	Log.Infof("Took %v to calculate", calcTime)

	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(ppd.AllTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.AllGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(ppd.CubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.PathBuilders)-1)

	start = time.Now()
	// force reload
	InitializePointDBEnv(env, true)
	loadTime := time.Now().Sub(start)
	Log.Infof("Took %v to load", loadTime)

	Log.Infof("Diff calc-load = %v", calcTime-loadTime)

	// Don't forget to get ppd different after init
	ppd, _ = GetServerPointPackData(env)
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.AllConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(ppd.AllTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.AllGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(ppd.CubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.PathBuilders)-1)
}

func TestSaveAll(t *testing.T) {
	m3db.Log.SetDebug()
	Log.SetDebug()
	m3util.SetToTestMode()

	tempEnv := GetCleanTempDb(m3util.PointTempEnv)
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
	loaded, _ := loadConnectionDetails(tempEnv)
	assert.Equal(t, ExpectedNbConns, len(loaded))

	// Init
	ppd.initConnections()

	// ************ Trio Details

	n, err = ppd.saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Should be able to run twice
	n, err = ppd.saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Test we can load
	loaded2 := ppd.loadTrioDetails()
	assert.Equal(t, ExpectedNbTrios, len(loaded2))

	// Init
	ppd.initTrioDetails()

	// ************ Growth Contexts

	n, err = ppd.saveAllGrowthContexts()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Should be able to run twice
	n, err = ppd.saveAllGrowthContexts()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Test we can load
	loaded3 := ppd.loadGrowthContexts()
	assert.Equal(t, ExpectedNbGrowthContexts, len(loaded3))

	// Init
	ppd.initGrowthContexts()

	// ************ Context Cubes

	n, err = ppd.saveAllContextCubes()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Should be able to run twice
	n, err = ppd.saveAllContextCubes()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Test we can load
	loaded4 := ppd.loadContextCubes()
	assert.Equal(t, ExpectedNbCubes, len(loaded4))

	// Init
	ppd.initContextCubes()

	// ************ Path Builders

	n, err = ppd.saveAllPathBuilders()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Should be able to run twice
	n, err = ppd.saveAllPathBuilders()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Test we can load
	loaded5 := ppd.loadPathBuilders()
	assert.Equal(t, ExpectedNbPathBuilders, len(loaded5)-1)

	// Init from Good DB
	ppd.initPathBuilders()
}
