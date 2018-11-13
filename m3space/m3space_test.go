package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type ExpectedSpaceState struct  {
	baseNodes int
	newNodes  int
}

func TestSingleRedEvent(t *testing.T) {
	DEBUG = true
	space := MakeSpace(3*9)

	InitConnectionDetails()
	assertEmptySpace(t, &space, 3*9)

	// Only latest counting
	space.EventOutgrowthThreshold = Distance(0)

	space.CreateSingleEventCenter()

	expectedState := map[TickTime]ExpectedSpaceState{
		0:{1, 0},
		1:{1, 3},
		4:{1, -2},
		5:{1, -10},
	}
	assertSpaceStates(t, &space, expectedState, 5)

	assertNearMainPoints(t, &space)
}

func assertEmptySpace(t *testing.T, space *Space, max int64) {
	assert.Equal(t, max, space.Max)
	assert.Equal(t, 0, len(space.nodesMap))
	assert.Equal(t, 0, len(space.connections))
	assert.Equal(t, 0, len(space.events))
	assert.Equal(t, 0, len(space.Elements))
}

func assertSpaceStates(t *testing.T, space *Space, expectMap map[TickTime]ExpectedSpaceState, finalTime TickTime) {
	expectedTime := TickTime(0)
	expect, ok := expectMap[expectedTime]
	assert.True(t, ok, "Should have the 0 tick time map entry in %v", expectMap)
	baseNodes := expect.baseNodes
	newNodes := baseNodes
	activeNodes := baseNodes
	nbNodes := baseNodes
	nbConnections := 0
	for {
		assertSpaceSingleEvent(t, space, expectedTime, nbNodes, nbConnections, activeNodes)
		if expectedTime == finalTime {
			break
		}
		space.ForwardTime()
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
		nbConnections += newNodes
		nbNodes += newNodes
	}
}

func assertSpaceSingleEvent(t *testing.T, space *Space, time TickTime, nbNodes, nbConnections int, nbActive int) {
	assert.Equal(t, time, space.currentTime)
	assert.Equal(t, nbNodes, len(space.nodesMap))
	assert.Equal(t, nbConnections, len(space.connections))
	assert.Equal(t, 1, len(space.events))
	assert.Equal(t, nbNodes+nbConnections+6, len(space.Elements))
	collectActiveElements := make([]*NodeDrawingElement, 0, 20)
	for _, draw := range space.Elements {
		if draw.Key() == NodeActive {
			nodeDrawing, ok := draw.(*NodeDrawingElement)
			assert.True(t, ok, "Node draw element should be of type NodeDrawingElement not %v", draw)
			collectActiveElements = append(collectActiveElements, nodeDrawing)
		}
	}
	assert.Equal(t, nbActive, len(collectActiveElements))
	for _, nodeDraw := range collectActiveElements {
		assert.Equal(t, uint8(1), nodeDraw.sdc.howManyColors())
	}
}

func assertNearMainPoints(t *testing.T, space *Space) {
	for _, node := range space.nodesMap {
		// Find main point attached to node
		var mainPointNode *Node
		if node.point.IsMainPoint() {
			mainPointNode = node
		} else {
			for _, conn := range node.connections {
				if conn != nil {
					if conn.N1.point.IsMainPoint() {
						mainPointNode = conn.N1
						break
					}
					if conn.N2.point.IsMainPoint() {
						mainPointNode = conn.N2
						break
					}
				}
			}
		}
		if mainPointNode != nil {
			assert.Equal(t, node.point.getNearMainPoint(), *(mainPointNode.point))
		}
	}
}
