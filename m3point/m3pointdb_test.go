package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var cleanedDbMutex sync.Mutex
var cleanedDb bool

func cleanDb() m3db.QsmEnvironment {
	env := m3db.TestEnv

	cleanedDbMutex.Lock()
	defer cleanedDbMutex.Unlock()

	if cleanedDb {
		return env
	}

	m3db.DropEnv(env)
	m3db.CheckOrCreateEnv(env)

	cleanedDb = true
	return env
}

func TestSaveAllConnections(t *testing.T) {
	Log.SetTrace()
	env := cleanDb()

	n, err := saveAllConnectionDetails(env)
	assert.Nil(t, err)
	assert.Equal(t, len(allConnections), n)
}
