package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var cleanedDbMutex sync.Mutex
var cleanedDb bool

func cleanDb() {
	pointEnv = m3db.GetEnvironment(m3db.TestEnv)

	cleanedDbMutex.Lock()
	defer cleanedDbMutex.Unlock()

	if cleanedDb {
		return
	}

	pointEnv.Destroy()

	pointEnv = m3db.GetEnvironment(m3db.TestEnv)
	cleanedDb = true
}

func TestSaveAllConnections(t *testing.T) {
	m3db.Log.SetTrace()
	Log.SetTrace()
	cleanDb()

	n, err := saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, len(allConnections), n)

	// Should be able to run twice
	n, err = saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, len(allConnections), n)

}
