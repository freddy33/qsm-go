package m3server

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
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
	m3db.SetEnvironmentCreator()

	if testDbFilled[envId] {
		return m3util.GetEnvironment(envId).(*m3db.QsmDbEnvironment)
	}

	m3util.RunQsm(envId, "run", "filldb")

	testDbFilled[envId] = true

	return m3util.GetEnvironment(envId).(*m3db.QsmDbEnvironment)
}

// Do not use this environment to load
func GetCleanTempDb(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use GetCleanTempDb in non test mode!")
	}
	m3db.SetEnvironmentCreator()
	env := m3util.GetEnvironment(envId).(*m3db.QsmDbEnvironment)

	dbMutex.Lock()
	defer dbMutex.Unlock()

	if cleanedDb[envId] {
		return env
	}

	env.Destroy()

	env = m3util.GetEnvironment(envId).(*m3db.QsmDbEnvironment)
	cleanedDb[envId] = true

	return env
}
