package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

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

func setFullTestDb() {
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

func TestSaveAllTrioContexts(t *testing.T) {
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

	n, err = saveAllTrioContexts()
	assert.Nil(t, err)
	assert.Equal(t, 52, n)

	// Should be able to run twice
	n, err = saveAllTrioContexts()
	assert.Nil(t, err)
	assert.Equal(t, 52, n)
}

func TestLoadOrCalculate(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()

	setFullTestDb()
	defer _nilPointEnv()

	start := time.Now()
	allConnections, allConnectionsByVector = calculateConnectionDetails()
	connectionsLoaded = true
	allTrioDetails = calculateAllTrioDetails()
	trioDetailsLoaded = true
	allTrioContexts = calculateAllTrioContexts()
	trioContextsLoaded = true
	allCubesPerContext = calculateAllContextCubes()
	cubesPerContextLoaded = true
	calcTime := time.Now().Sub(start)
	Log.Infof("Took %v to calculate", calcTime)

	assert.Equal(t, 50, len(allConnections))
	assert.Equal(t, 50, len(allConnectionsByVector))
	assert.Equal(t, 200, len(allTrioDetails))
	assert.Equal(t, 52, len(allTrioContexts))
	assert.Equal(t, 52, len(allCubesPerContext))

	start = time.Now()
	// force reload
	connectionsLoaded = false
	trioDetailsLoaded = false
	trioContextsLoaded = false
	cubesPerContextLoaded = false
	Initialize()
	loadTime := time.Now().Sub(start)
	Log.Infof("Took %v to load", loadTime)

	Log.Infof("Diff calc-load = %v", calcTime - loadTime)

	assert.Equal(t, 50, len(allConnections))
	assert.Equal(t, 50, len(allConnectionsByVector))
	assert.Equal(t, 200, len(allTrioDetails))
	assert.Equal(t, 52, len(allTrioContexts))
	assert.Equal(t, 52, len(allCubesPerContext))
}
