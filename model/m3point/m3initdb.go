package m3point

import (
	"github.com/freddy33/qsm-go/utils/m3db"
	"strconv"
	"sync"
	"time"
)

func InitializeDBEnv(env *m3db.QsmEnvironment, forced bool) {
	ppd := GetPointPackData(env)
	if forced {
		ppd.resetFlags()
	}
	ppd.initConnections()
	ppd.initTrioDetails()
	ppd.initGrowthContexts()
	ppd.initContextCubes()
	ppd.initPathBuilders()
}

func (ppd *PointPackData) initConnections() {
	if !ppd.ConnectionsLoaded {
		ppd.AllConnections, ppd.AllConnectionsByVector = loadConnectionDetails(ppd.Env)
		ppd.ConnectionsLoaded = true
		Log.Debugf("Environment %d has %d connection details", ppd.GetEnvId(), len(ppd.AllConnections))
	}
}

func (ppd *PointPackData) initTrioDetails() {
	if !ppd.TrioDetailsLoaded {
		ppd.AllTrioDetails = ppd.loadTrioDetails()
		ppd.TrioDetailsLoaded = true
		Log.Debugf("Environment %d has %d trio details", ppd.GetEnvId(), len(ppd.AllTrioDetails))
	}
}

func (ppd *PointPackData) initGrowthContexts() {
	if !ppd.GrowthContextsLoaded {
		ppd.AllGrowthContexts = ppd.loadGrowthContexts()
		ppd.GrowthContextsLoaded = true
		Log.Debugf("Environment %d has %d growth contexts", ppd.GetEnvId(), len(ppd.AllGrowthContexts))
	}
}

func (ppd *PointPackData) initContextCubes() {
	if !ppd.CubesLoaded {
		ppd.CubeIdsPerKey = ppd.loadContextCubes()
		ppd.CubesLoaded = true
		Log.Debugf("Environment %d has %d cubes", ppd.GetEnvId(), len(ppd.CubeIdsPerKey))
	}
}

func (ppd *PointPackData) initPathBuilders() {
	if !ppd.PathBuildersLoaded {
		ppd.PathBuilders = ppd.loadPathBuilders()
		ppd.PathBuildersLoaded = true
		Log.Debugf("Environment %d has %d path builders", ppd.GetEnvId(), len(ppd.PathBuilders))
	}
}

func ReFillDbEnv(env *m3db.QsmEnvironment) {
	env.Destroy()
	time.Sleep(1000 * time.Millisecond)
	FillDbEnv(env)
}

func FillDbEnv(env *m3db.QsmEnvironment) {
	ppd := GetPointPackData(env)

	n, err := ppd.saveAllConnectionDetails()
	if err != nil {
		Log.Fatalf("could not save all connections due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d connection details", ppd.GetEnvId(), n)
	}

	ppd.initConnections()

	n, err = ppd.saveAllTrioDetails()
	if err != nil {
		Log.Fatalf("could not save all trios due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio details", ppd.GetEnvId(), n)
	}
	ppd.initTrioDetails()

	n, err = ppd.saveAllGrowthContexts()
	if err != nil {
		Log.Fatalf("could not save all growth contexts due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d growth contexts", ppd.GetEnvId(), n)
	}
	ppd.initGrowthContexts()

	n, err = ppd.saveAllContextCubes()
	if err != nil {
		Log.Fatalf("could not save all contexts cubes due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d contexts cubes", ppd.GetEnvId(), n)
	}
	ppd.initContextCubes()

	n, err = ppd.saveAllPathBuilders()
	if err != nil {
		Log.Fatalf("could not save all path builders due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d path builders", ppd.GetEnvId(), n)
	}
	ppd.initPathBuilders()
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

	if testDbFilled[envId] {
		return m3db.GetEnvironment(envId)
	}

	envNumber := strconv.Itoa(int(envId))

	m3db.FillDb(envNumber)

	testDbFilled[envId] = true

	return m3db.GetEnvironment(envId)
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
