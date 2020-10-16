package m3gl

import (
	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ExpectedSpaceState struct {
	baseNodes int
	newNodes  int
}

func getGlTestEnv() *client.QsmApiEnvironment {
	return client.GetOrCreateInitializedApiEnv(m3util.GlTestEnv, true, true)
}

func TestSingleRedEvent(t *testing.T) {
	Log.SetDebug()
	m3space.Log.SetDebug()
	m3util.SetToTestMode()

	max := m3space.MinMaxCoord
	world := MakeWorld(getGlTestEnv(), max, 0.0)

	//assertEmptyWorld(t, &world, max)

	_, err := world.WorldSpace.CreateEvent(m3point.GrowthType(8), 0, 0, m3space.ZeroDistAndTime, m3point.Origin, m3space.RedEvent)
	if !assert.NoError(t, err) {
		return
	}
	world.CreateDrawingElements()

	expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
		0: {1, 0},
		1: {1, 3},
		4: {1, -2},
		5: {1, -10},
	}
	assertSpaceStates(t, &world, expectedState, 5)
}

func assertEmptyWorld(t *testing.T, world *DisplayWorld, max m3point.CInt) {
	assert.Equal(t, max, world.WorldSpace.GetMaxCoord())
	spaceTime := world.GetSpaceTime()
	assert.Nil(t, spaceTime)
	assert.Equal(t, 0, len(world.Elements))
}

func assertSpaceStates(t *testing.T, world *DisplayWorld, expectMap map[m3space.DistAndTime]ExpectedSpaceState, finalTime m3space.DistAndTime) {
	expectedTime := m3space.ZeroDistAndTime
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
		world.CreateDrawingElements()
		expectedTime++
		expect, ok = expectMap[expectedTime]
		if ok {
			if expect.newNodes < 0 {
				newNodes *= 2
				newNodes += expect.newNodes
			} else {
				newNodes = expect.newNodes
			}
		} else {
			newNodes *= 2
		}
		activeNodes = newNodes + baseNodes
		nbNodes += newNodes
		nbConnections += newNodes
	}
}

func assertSpaceSingleEvent(t *testing.T, world *DisplayWorld, time m3space.DistAndTime, nbNodes, nbConnections int, nbActive int) {
	spaceTime := world.GetSpaceTime()
	assert.Equal(t, time, spaceTime.GetCurrentTime(), "failed at %d", time)
	assert.Equal(t, nbActive, spaceTime.GetNbActiveNodes(), "failed at %d", time)
	// TODO: Change all test to use real active links when both sides are active
	//assert.Equal(t, nbConnections, world.WorldSpace.GetNbActiveLinks(), "failed at %d", time)
	assert.Equal(t, 1, len(world.WorldSpace.GetActiveEventsAt(0)), "failed at %d", time)
	assert.Equal(t, spaceTime.GetNbActiveNodes()+spaceTime.GetNbActiveLinks()+6, len(world.Elements), "failed at %d", time)
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
