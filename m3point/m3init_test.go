package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func SetFullTestDb() {
	envNumber := strconv.Itoa(int(m3db.TestEnv))
	origQsmId := os.Getenv(m3db.QsmEnvNumberKey)

	if envNumber != origQsmId {
		// Reset the env var to what it was on exit of this method
		defer m3db.SetEnvQuietly(m3db.QsmEnvNumberKey, origQsmId)
		// set the env var correctly
		m3util.ExitOnError(os.Setenv(m3db.QsmEnvNumberKey, envNumber))
	}

	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "run", "filldb")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Fatalf("failed to fill db for test environment %s at OS level due to %v with output: ***\n%s\n***", envNumber, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("check environment %s at OS output: ***\n%s\n***", envNumber, string(out))
		}
	}

	pointEnv = m3db.GetEnvironment(m3db.TestEnv)
}

func TestLoadOrCalculate(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()

	SetFullTestDb()

	start := time.Now()
	allConnections, allConnectionsByVector = calculateConnectionDetails()
	connectionsLoaded = true
	allTrioDetails = calculateAllTrioDetails()
	trioDetailsLoaded = true
	allGrowthContexts = calculateAllGrowthContexts()
	growthContextsLoaded = true
	cubeIdsPerKey = calculateAllContextCubes()
	cubesLoaded = true
	pathBuilders = calculateAllPathBuilders()
	pathBuildersLoaded = true
	calcTime := time.Now().Sub(start)
	Log.Infof("Took %v to calculate", calcTime)

	assert.Equal(t, 50, len(allConnections))
	assert.Equal(t, 50, len(allConnectionsByVector))
	assert.Equal(t, 200, len(allTrioDetails))
	assert.Equal(t, 52, len(allGrowthContexts))
	assert.Equal(t, 5192, len(cubeIdsPerKey))
	assert.Equal(t, 5192 + 1, len(pathBuilders))

	start = time.Now()
	// force reload
	connectionsLoaded = false
	trioDetailsLoaded = false
	growthContextsLoaded = false
	cubesLoaded = false
	pathBuildersLoaded = false
	Initialize()
	loadTime := time.Now().Sub(start)
	Log.Infof("Took %v to load", loadTime)

	Log.Infof("Diff calc-load = %v", calcTime - loadTime)

	assert.Equal(t, 50, len(allConnections))
	assert.Equal(t, 50, len(allConnectionsByVector))
	assert.Equal(t, 200, len(allTrioDetails))
	assert.Equal(t, 52, len(allGrowthContexts))
	assert.Equal(t, 5192, len(cubeIdsPerKey))
	assert.Equal(t, 5192 + 1, len(pathBuilders))
}

/* TODO: Reactivate once better concurrent system on DB

var cleanedDbMutex sync.Mutex
var cleanedDb bool

func setCleanTempDb() {
	pointEnv = m3db.GetEnvironment(m3db.TempEnv)

	cleanedDbMutex.Lock()
	defer cleanedDbMutex.Unlock()

	if cleanedDb {
		return
	}

	pointEnv.Destroy()

	pointEnv = m3db.GetEnvironment(m3db.TempEnv)
	cleanedDb = true
}

func _closePointEnv() {
	defer _nilPointEnv()
	m3db.CloseEnv(pointEnv)
}

func _nilPointEnv() {
	pointEnv = nil
}

func TestSaveAllConnections(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()

	setCleanTempDb()
	defer _closePointEnv()

	n, err := saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, 50, n)

	// Should be able to run twice
	n, err = saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, 50, n)
}

func TestSaveAllTrios(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()

	setCleanTempDb()
	defer _closePointEnv()

	n, err := saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, 50, n)

	// Init from DB
	initConnections()

	n, err = saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, 200, n)

	// Should be able to run twice
	n, err = saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, 200, n)
}

func TestSaveAllGrowthContexts(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()

	setCleanTempDb()
	defer _closePointEnv()

	n, err := saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, 50, n)
	initConnections()

	n, err = saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, 200, n)
	initTrioDetails()

	n, err = saveAllGrowthContexts()
	assert.Nil(t, err)
	assert.Equal(t, 52, n)

	// Should be able to run twice
	n, err = saveAllGrowthContexts()
	assert.Nil(t, err)
	assert.Equal(t, 52, n)
}

*/
