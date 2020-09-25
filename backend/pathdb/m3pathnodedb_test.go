package pathdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathNodeDbConnMask(t *testing.T) {
	m3util.SetToTestMode()

	pn := getNewPathNodeDb()
	assert.Equal(t, NewPathNode, pn.state)
	assert.Equal(t, uint16(m3path.ConnectionNotSet), pn.connectionMask)

	assert.True(t, pn.IsNew())
	assert.False(t, pn.IsInPool())
	for i := 0; i < m3path.NbConnections; i++ {
		assert.Equal(t, m3path.ConnectionNotSet, pn.GetConnectionState(i))
	}

	pn.SetConnectionState(1, m3path.ConnectionFrom)
	if pn.state == SyncInDbPathNode {
		pn.state = ModifiedNode
	}
	assert.Equal(t, m3path.ConnectionNotSet, pn.GetConnectionState(0))
	assert.Equal(t, m3path.ConnectionFrom, pn.GetConnectionState(1))
	assert.Equal(t, m3path.ConnectionNotSet, pn.GetConnectionState(2))

	pn.release()
	assert.False(t, pn.IsNew())
	assert.True(t, pn.IsInPool())
}
