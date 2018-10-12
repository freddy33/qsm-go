package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSpace(t *testing.T) {
	assert.Equal(t, int64(9), SpaceObj.max)
	assert.Equal(t, 0, len(SpaceObj.nodes))
	assert.Equal(t, 0, len(SpaceObj.connections))
	assert.Equal(t, 0, len(SpaceObj.events))
	assert.Equal(t, 0, len(SpaceObj.Elements))

	SpaceObj.CreateStuff(3)

	assert.Equal(t, int64(3), SpaceObj.max)
	// Big nodes = (center + center face + middle edge + corner) * (main + 3)
	nbNodes := (1 + 6 + 12 + 8) * 4
	assert.Equal(t, nbNodes, len(SpaceObj.nodes))
	assert.Equal(t, nbNodes+35, len(SpaceObj.connections))
	assert.Equal(t, 4, len(SpaceObj.events))
	assert.Equal(t, 442, len(SpaceObj.Elements))
	assert.Equal(t, TickTime(0), SpaceObj.currentTime)
	assertOutgrowth(t, 4)
	assertOutgrowthDistance(t,1)

	SpaceObj.ForwardTime()
	// Same elements just color changes
	assert.Equal(t, nbNodes, len(SpaceObj.nodes))
	assert.Equal(t, nbNodes+35, len(SpaceObj.connections))
	assert.Equal(t, 4, len(SpaceObj.events))
	assert.Equal(t, 442, len(SpaceObj.Elements))
	assert.Equal(t, TickTime(1), SpaceObj.currentTime)
	assertOutgrowth(t, 4+(4*3))
	assertOutgrowthDistance(t,3)
}

func assertOutgrowth(t *testing.T, expect int) {
	nbOutgrowth := 0
	for _, evt := range SpaceObj.events {
		nbOutgrowth += len(evt.O)
	}
	assert.Equal(t, expect, nbOutgrowth)
	nbOutgrowth = 0
	for _, node := range SpaceObj.nodes {
		nbOutgrowth += len(node.E)
	}
	assert.Equal(t, expect, nbOutgrowth)
}

func assertOutgrowthDistance(t *testing.T, topOnes int) {
	for _, evt := range SpaceObj.events {
		nbTopOnes := 0
		for _, eo := range evt.O {
			if eo.D == Distance(SpaceObj.currentTime-evt.T) {
				nbTopOnes++
			}
		}
		assert.Equal(t, topOnes, nbTopOnes, "NB top ones expected failed for", evt.ID)
	}
}
