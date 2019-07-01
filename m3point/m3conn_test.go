package m3point

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnectionDetails(t *testing.T) {
	Log.SetInfo()
	Initialize()

	for k, v := range allConnectionsByVector {
		assert.Equal(t, k, v.Vector)
		assert.Equal(t, k.DistanceSquared(), v.DistanceSquared())
		currentNumber := v.GetPosId()
		sameNumber := 0
		for _, nv := range allConnectionsByVector {
			if nv.GetPosId() == currentNumber {
				sameNumber++
				if nv.Vector != v.Vector {
					assert.Equal(t, nv.GetId(), -v.GetId(), "Should have opposite id")
					assert.Equal(t, nv.Vector.Neg(), v.Vector, "Should have neg vector")
				}
			}
		}
		assert.Equal(t, 2, sameNumber, "Should have 2 with same conn number for %d", currentNumber)
	}

	countConnId := make(map[ConnectionId]int)
	for i, tA := range allBaseTrio {
		for j, tB := range allBaseTrio {
			connVectors := GetNonBaseConnections(tA, tB)
			for k, connVector := range connVectors {
				connDetails, ok := allConnectionsByVector[connVector]
				assert.True(t, ok, "Connection between 2 trio (%d,%d) number %k is not in conn details", i, j, k)
				assert.Equal(t, connVector, connDetails.Vector, "Connection between 2 trio (%d,%d) number %k is not in conn details", i, j, k)
				countConnId[connDetails.GetId()]++
			}
		}
	}
	Log.Debug("ConnId usage:", countConnId)
}
