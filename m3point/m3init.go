package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var pointEnvId m3db.QsmEnvID
var pointEnv *m3db.QsmEnvironment

var connectionsLoaded [m3db.MaxNumberOfEnvironments]bool
var trioDetailsLoaded [m3db.MaxNumberOfEnvironments]bool
var growthContextsLoaded [m3db.MaxNumberOfEnvironments]bool
var cubesLoaded [m3db.MaxNumberOfEnvironments]bool
var pathBuildersLoaded [m3db.MaxNumberOfEnvironments]bool

func GetPointEnv() *m3db.QsmEnvironment {
	if pointEnv == nil || pointEnv.GetConnection() == nil {
		if pointEnvId == m3db.NoEnv {
			pointEnvId = m3db.GetDefaultEnvId()
		}
		pointEnv = m3db.GetEnvironment(pointEnvId)
	}
	return pointEnv
}

func checkConnInitialized() {
	if !connectionsLoaded[pointEnvId] {
		Log.Fatal("Connections should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkTrioInitialized() {
	if !trioDetailsLoaded[pointEnvId] {
		Log.Fatal("Trios should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkGrowthContextsInitialized() {
	if !trioDetailsLoaded[pointEnvId] {
		Log.Fatal("trio contexts should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkCubesInitialized() {
	if !cubesLoaded[pointEnvId] {
		Log.Fatal("Cubes should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkPathBuildersInitialized() {
	if !pathBuildersLoaded[pointEnvId] {
		Log.Fatal("Path Builders should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func Initialize() {
	env := GetPointEnv()
	InitializeEnv(env, false)
	pointEnvId = env.GetId()
}

func InitializeEnv(env *m3db.QsmEnvironment, forced bool) {
	if forced {
		envId := env.GetId()
		connectionsLoaded[envId] = false
		trioDetailsLoaded[envId] = false
		growthContextsLoaded[envId] = false
		cubesLoaded[envId] = false
		pathBuildersLoaded[envId] = false
	}
	initConnections(env)
	initTrioDetails(env)
	initGrowthContexts(env)
	initContextCubes(env)
	initPathBuilders(env)
}

func initConnections(env *m3db.QsmEnvironment) {
	if !connectionsLoaded[env.GetId()] {
		allConnections, allConnectionsByVector = loadConnectionDetails(env)
		connectionsLoaded[env.GetId()] = true
		Log.Debugf("Environment %d has %d connection details", env.GetId(), len(allConnections))
	}
}

func initTrioDetails(env *m3db.QsmEnvironment) {
	if !trioDetailsLoaded[env.GetId()] {
		allTrioDetails = loadTrioDetails(env)
		trioDetailsLoaded[env.GetId()] = true
		Log.Debugf("Environment %d has %d trio details", env.GetId(), len(allTrioDetails))
	}
}

func initGrowthContexts(env *m3db.QsmEnvironment) {
	if !growthContextsLoaded[env.GetId()] {
		allGrowthContexts = loadGrowthContexts(env)
		growthContextsLoaded[env.GetId()] = true
		Log.Debugf("Environment %d has %d growth contexts", env.GetId(), len(allGrowthContexts))
	}
}

func initContextCubes(env *m3db.QsmEnvironment) {
	if !cubesLoaded[env.GetId()] {
		cubeIdsPerKey = loadContextCubes(env)
		cubesLoaded[env.GetId()] = true
		Log.Debugf("Environment %d has %d cubes", env.GetId(), len(cubeIdsPerKey))
	}
}

func initPathBuilders(env *m3db.QsmEnvironment) {
	if !pathBuildersLoaded[env.GetId()] {
		pathBuilders = loadPathBuilders(env)
		pathBuildersLoaded[env.GetId()] = true
		Log.Debugf("Environment %d has %d path builders", env.GetId(), len(pathBuilders))
	}
}

func ReFillDb() {
	ReFillDbEnv(GetPointEnv())
}

func ReFillDbEnv(env *m3db.QsmEnvironment) {
	env.Destroy()
	pointEnv = nil
	time.Sleep(100 * time.Millisecond)
	FillDb()
}

func FillDb() {
	FillDbEnv(GetPointEnv())
}

func FillDbEnv(env *m3db.QsmEnvironment) {
	n, err := saveAllConnectionDetails(env)
	if err != nil {
		Log.Fatalf("could not save all connections due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d connection details", env.GetId(), n)
	}

	initConnections(env)

	n, err = saveAllTrioDetails(env)
	if err != nil {
		Log.Fatalf("could not save all trios due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio details", env.GetId(), n)
	}
	initTrioDetails(env)

	n, err = saveAllGrowthContexts(env)
	if err != nil {
		Log.Fatalf("could not save all trio contexts due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio contexts", env.GetId(), n)
	}
	initGrowthContexts(env)

	n, err = saveAllContextCubes(env)
	if err != nil {
		Log.Fatalf("could not save all contexts cubes due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d contexts cubes", env.GetId(), n)
	}
	initContextCubes(env)

	n, err = saveAllPathBuilders(env)
	if err != nil {
		Log.Fatalf("could not save all path builders due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d path builders", env.GetId(), n)
	}
	initPathBuilders(env)
}

/***************************************************************/
// Utility methods for test
/***************************************************************/

var dbMutex sync.Mutex
var cleanedDb [m3db.MaxNumberOfEnvironments]bool
var testDbFilled [m3db.MaxNumberOfEnvironments]bool

func GetFullTestDb(envId m3db.QsmEnvID) *m3db.QsmEnvironment {
	if !m3db.TestMode {
		Log.Fatalf("Cannot use GetFullTestDb in non test mode!")
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	pointEnvId = envId

	if testDbFilled[envId] {
		pointEnv = nil
		return GetPointEnv()
	}

	envNumber := strconv.Itoa(int(pointEnvId))
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

	testDbFilled[envId] = true

	return GetPointEnv()
}

// Do not use this environment to load
func GetCleanTempDb(envId m3db.QsmEnvID) *m3db.QsmEnvironment {
	if !m3db.TestMode {
		Log.Fatalf("Cannot use GetCleanTempDb in non test mode!")
	}

	env := m3db.GetEnvironment(envId)

	dbMutex.Lock()
	defer dbMutex.Unlock()

	if cleanedDb[envId] {
		return env
	}

	env.Destroy()

	env = m3db.GetEnvironment(envId)
	cleanedDb[envId] = true

	return env
}

