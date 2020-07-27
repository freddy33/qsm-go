package m3server

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

func getServerFullTestDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use getServerFullTestDb in non test mode!")
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	if testDbFilled[envId] {
		return m3db.GetEnvironment(envId)
	}

	m3util.RunQsm(envId, "run", "filldb")

	testDbFilled[envId] = true

	return m3db.GetEnvironment(envId)
}

// Do not use this environment to load
func GetCleanTempDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	if !m3util.TestMode {
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
