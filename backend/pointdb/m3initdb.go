package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

/***************************************************************/
// Utility methods for test
/***************************************************************/

var dbMutex sync.Mutex

var cleanedDb [m3util.MaxNumberOfEnvironments]bool
var testDbFilled [m3util.MaxNumberOfEnvironments]bool

func GetServerFullTestDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use getServerFullTestDb in non test mode!")
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	env := m3db.GetEnvironment(envId)

	if testDbFilled[envId] {
		return m3db.GetEnvironment(envId)
	}

	err := env.CheckSchema()
	if err != nil {
		Log.Fatal(err)
	}
	FillDbEnv(env)

	testDbFilled[envId] = true

	InitializePointDBEnv(env, false)
	return env
}

// Do not use this environment to load
func GetCleanTempDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use GetCleanTempDb in non test mode!")
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	env := m3db.GetEnvironment(envId)

	if cleanedDb[envId] {
		return env
	}
	env.Destroy()
	cleanedDb[envId] = true

	env = m3db.GetEnvironment(envId)
	err := env.CheckSchema()
	if err != nil {
		Log.Fatal(err)
	}
	return env
}
