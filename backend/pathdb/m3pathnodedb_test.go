package pathdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathNodeDbConnMask(t *testing.T) {
	m3util.SetToTestMode()
	env := GetFullTestDb(m3util.PathTestEnv)
	InitializePathDBEnv(env)

	pn := getNewPathNodeDb()
	assert.Equal(t, NewPathNode, pn.state)
	assert.Equal(t, uint16(ConnectionNotSet), pn.connectionMask)

	assert.True(t, pn.IsNew())
	assert.False(t, pn.IsInPool())
	for i := 0; i < NbConnections; i++ {
		assert.Equal(t, ConnectionNotSet, pn.getConnectionState(i))
	}

	pn.setConnectionState(1, ConnectionFrom)
	assert.Equal(t, ConnectionNotSet, pn.getConnectionState(0))
	assert.Equal(t, ConnectionFrom, pn.getConnectionState(1))
	assert.Equal(t, ConnectionNotSet, pn.getConnectionState(2))

	pn.release()
	assert.False(t, pn.IsNew())
	assert.True(t, pn.IsInPool())
}
