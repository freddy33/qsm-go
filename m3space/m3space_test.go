package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSpace(t *testing.T) {
	assert.Equal(t, int64(9), SpaceObj.max)
	assert.Equal(t, 0, len(SpaceObj.nodes))
	assert.Equal(t, 0, len(SpaceObj.connections))
	assert.Equal(t, 0, len(SpaceObj.Elements))

	SpaceObj.CreateStuff(3)

	assert.Equal(t, int64(3), SpaceObj.max)
	// Big nodes = (center + center face + middle edge + corner) * (main + 3)
	nbNodes := (1 + 6 + 12 + 8) * 4
	assert.Equal(t, nbNodes, len(SpaceObj.nodes))
	assert.Equal(t, nbNodes+35, len(SpaceObj.connections))
	assert.Equal(t, nbNodes*4+7, len(SpaceObj.Elements))
}
