package m3gl

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/freddy33/qsm-go/m3space"
)

type ExpectedSpaceState struct {
	baseNodes int
	newNodes  int
}

func TestSingleRedEvent(t *testing.T) {
	DEBUG = true
	world := MakeWorld(3*9, 0.0)

	assertEmptyWorld(t, &world, 3*9)

	// Only latest counting
	world.Space.EventOutgrowthThreshold = m3space.Distance(0)

	world.Space.CreateSingleEventCenter()
	world.CreateDrawingElements()

	expectedState := map[m3space.TickTime]ExpectedSpaceState{
		0: {1, 0},
		1: {1, 3},
		4: {1, -2},
		5: {1, -10},
	}
	assertSpaceStates(t, &world, expectedState, 5)
}

func assertEmptyWorld(t *testing.T, world *DisplayWorld, max int64) {
	assert.Equal(t, max, world.Space.Max)
	assert.Equal(t, 0, world.Space.GetNbNodes())
	assert.Equal(t, 0, world.Space.GetNbConnections())
	assert.Equal(t, 0, world.Space.GetNbEvents())
	assert.Equal(t, 0, len(world.Elements))
}

func assertSpaceStates(t *testing.T, world *DisplayWorld, expectMap map[m3space.TickTime]ExpectedSpaceState, finalTime m3space.TickTime) {
	expectedTime := m3space.TickTime(0)
	expect, ok := expectMap[expectedTime]
	assert.True(t, ok, "Should have the 0 tick time map entry in %v", expectMap)
	baseNodes := expect.baseNodes
	newNodes := baseNodes
	activeNodes := baseNodes
	nbNodes := baseNodes
	nbConnections := 0
	for {
		assertSpaceSingleEvent(t, world, expectedTime, nbNodes, nbConnections, activeNodes)
		if expectedTime == finalTime {
			break
		}
		world.ForwardTime()
		expectedTime++
		expect, ok = expectMap[expectedTime]
		if ok {
			if expect.newNodes < 0 {
				newNodes *= 2
				nbConnections += newNodes
				newNodes += expect.newNodes
			} else {
				newNodes = expect.newNodes
				nbConnections += newNodes
			}
		} else {
			newNodes *= 2
			nbConnections += newNodes
		}
		activeNodes = newNodes + baseNodes
		nbNodes += newNodes
	}
}

func assertSpaceSingleEvent(t *testing.T, world *DisplayWorld, time m3space.TickTime, nbNodes, nbConnections int, nbActive int) {
	assert.Equal(t, time, world.Space.GetCurrentTime(), "failed at %d", time)
	assert.Equal(t, nbNodes, world.Space.GetNbNodes(), "failed at %d", time)
	assert.Equal(t, nbConnections, world.Space.GetNbConnections(), "failed at %d", time)
	assert.Equal(t, 1, world.Space.GetNbEvents(), "failed at %d", time)
	assert.Equal(t, nbNodes+nbConnections+6, len(world.Elements), "failed at %d", time)
	nbDisplay := 0
	collectActiveElements := make([]*NodeDrawingElement, 0, 20)
	for _, draw := range world.Elements {
		if draw.Key() == NodeActive {
			nodeDrawing, ok := draw.(*NodeDrawingElement)
			assert.True(t, ok, "Node draw element should be of type NodeDrawingElement not %v", draw)
			collectActiveElements = append(collectActiveElements, nodeDrawing)
		}
		if draw.Display(world.Filter) {
			// TODO: Should be able to test active connections...
			if !draw.Key().IsConnection() {
				nbDisplay++
			}
		}
	}
	assert.Equal(t, 6+nbActive, nbDisplay, "failed at %d", time)
	assert.Equal(t, nbActive, len(collectActiveElements), "failed at %d", time)
	for _, nodeDraw := range collectActiveElements {
		assert.Equal(t, uint8(1), nodeDraw.sdc.howManyColors(), "failed at %d", time)
	}
}
