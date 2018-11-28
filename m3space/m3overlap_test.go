package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestOverlapSameEvent(t *testing.T) {
	Log.Level = m3util.TRACE
	space := MakeSpace(3 * 9)

	InitConnectionDetails()
	assertEmptySpace(t, &space, 3*9)

	// Only latest counting
	space.SetEventOutgrowthThreshold(Distance(0))
	ctx := GrowthContext{&Origin, 1, 0, false, 0}
	space.CreateEventWithGrowthContext(Origin, RedEvent, ctx)

	expectedTime := TickTime(0)
	nbLatestNodes := 1
	// No overlap until time 5
	for expectedTime < 5 {
		assert.Equal(t, expectedTime, space.currentTime)
		latestNodes := getAllNodeWithOutgrowthAtD(t, &space, Distance(expectedTime))
		assert.Equal(t, nbLatestNodes, len(latestNodes), "nbLatestNodes failed at %d", expectedTime)
		space.ForwardTime()
		expectedTime++
		if expectedTime == 1 {
			nbLatestNodes = 3
		} else {
			nbLatestNodes *= 2
		}
	}

	assert.Equal(t, TickTime(5), expectedTime)
	assert.Equal(t, expectedTime, space.currentTime)
	assert.Equal(t, 3*2*2*2*2, nbLatestNodes)

	latestNodes := getAllNodeWithOutgrowthAtD(t, &space, Distance(expectedTime))

	assert.Equal(t, nbLatestNodes-13, len(latestNodes))
}

// Retrieve all nodes having outgrowth at exact distance d from the event
func getAllNodeWithOutgrowthAtD(t *testing.T, space *Space, atD Distance) map[Point]*Node {
	res := make(map[Point]*Node, 25)
	for _, evt := range space.events {
		for _, eo := range evt.latestOutgrowths {
			if eo.distance == atD {
				res[*eo.pos] = space.GetNode(*eo.pos)
			}
		}
		for _, eo := range evt.currentOutgrowths {
			if eo.distance == atD {
				res[*eo.pos] = space.GetNode(*eo.pos)
			}
		}
	}
	return res
}
