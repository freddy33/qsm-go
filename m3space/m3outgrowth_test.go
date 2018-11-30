package m3space

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOverlapSameEvent(t *testing.T) {
	LogStat.Level = m3util.WARN
	Log.Level = m3util.TRACE
	space := MakeSpace(3 * 9)

	assertEmptySpace(t, &space, 3*9)

	// Only latest counting
	space.SetEventOutgrowthThreshold(Distance(0))
	space.blockOnSameEvent = 4
	ctx := GrowthContext{&Origin, 1, 0, false, 0}
	space.CreateEventWithGrowthContext(Origin, RedEvent, ctx)

	expectedTime := TickTime(0)
	nbLatestNodes := 1
	// No overlap until time 5
	for expectedTime < 5 {
		assert.Equal(t, expectedTime, space.currentTime)
		latestNodes := getAllNodeWithOutgrowthAtD(&space, Distance(expectedTime))
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

	latestNodes := getAllNodeWithOutgrowthAtD(&space, Distance(expectedTime))

	assert.Equal(t, nbLatestNodes-13, len(latestNodes))
	// Single vent means only one latest outgrowth per active point
	latestOutgrowths := make(map[Point]Outgrowth, len(latestNodes))
	fromSizeHisto := make(map[int]int, 3)
	for _, evt := range space.events {
		for _, eo := range evt.latestOutgrowths {
			fromSizeHisto[eo.FromLength()]++
			lo, ok := latestOutgrowths[*eo.pos]
			if ok {
				assert.Fail(t, "Should not have an outgrowth at %v for %v <= %v", *eo.pos, eo.String(), lo)
			} else {
				latestOutgrowths[*eo.pos] = eo
			}
		}
	}
	assert.Equal(t, nbLatestNodes-13, len(latestOutgrowths))
	Log.Info("From size histo", fromSizeHisto)
}

// Retrieve all nodes having outgrowth at exact distance d from the event
func getAllNodeWithOutgrowthAtD(space *Space, atD Distance) map[Point]Node {
	res := make(map[Point]Node, 25)
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
