package m3gl

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/freddy33/qsm-go/utils/m3util"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type ExpectedSpaceState struct {
	baseNodes int
	newNodes  int
}

var envMutex sync.Mutex

func getGlTestEnv() *m3db.QsmDbEnvironment {
	envMutex.Lock()
	defer envMutex.Unlock()
	m3util.SetToTestMode()
	glEnv := m3path.GetFullTestDb(m3util.GlTestEnv)
	m3point.InitializeDBEnv(glEnv, true)
	return glEnv
}

func TestSingleRedEvent(t *testing.T) {
	Log.SetDebug()
	m3space.Log.SetDebug()
	m3util.SetToTestMode()

	m3path.GetFullTestDb(m3util.GlTestEnv)

	world := MakeWorld(getGlTestEnv(), 3*9, 0.0)

	assertEmptyWorld(t, &world, 3*9)

	// Only latest counting
	world.WorldSpace.SetEventOutgrowthThreshold(m3space.DistAndTime(0))

	world.WorldSpace.CreateSingleEventCenter()
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
	assert.Equal(t, max, world.WorldSpace.Max)
	assert.Equal(t, 0, world.WorldSpace.GetNbNodes())
	assert.Equal(t, 0, world.WorldSpace.GetNbActiveLinks())
	assert.Equal(t, 0, world.WorldSpace.GetNbEvents())
	assert.Equal(t, 0, len(world.Elements))
}

func assertSpaceStates(t *testing.T, world *DisplayWorld, expectMap map[m3space.DistAndTime]ExpectedSpaceState, finalTime m3space.DistAndTime) {
	expectedTime := m3space.DistAndTime(0)
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
	assert.Equal(t, time, world.WorldSpace.GetCurrentTime(), "failed at %d", time)
	assert.Equal(t, nbNodes, world.WorldSpace.GetNbNodes(), "failed at %d", time)
	// TODO: Change all test to use real active links when both sides are active
	//assert.Equal(t, nbConnections, world.WorldSpace.GetNbActiveLinks(), "failed at %d", time)
	assert.Equal(t, 1, world.WorldSpace.GetNbEvents(), "failed at %d", time)
	assert.Equal(t, world.WorldSpace.GetNbActiveNodes()+world.WorldSpace.GetNbActiveLinks()+6, len(world.Elements), "failed at %d", time)
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
