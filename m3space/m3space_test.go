package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ExpectedSpaceState struct {
	baseNodes      int
	newNodes       int
	oldActiveNodes int
}

func TestSingleRedEventD0(t *testing.T) {
	Log.Level = m3util.INFO
	InitConnectionDetails()
	for trioIdx := 0; trioIdx < 12; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))

		evt := space.CreateSingleEventCenter()
		evt.growthContext.permutationIndex = trioIdx

		deltaT8FromIdx0 := 0
		deltaT9FromIdx0 := 0
		deltaT10FromIdx0 := 0
		deltaT11FromIdx0 := 0
		if AllMod8Permutations[trioIdx][3] != 5 {
			deltaT8FromIdx0 = 5
			deltaT9FromIdx0 = -11
			deltaT10FromIdx0 = -13
			deltaT11FromIdx0 = -22
		}

		expectedState := map[TickTime]ExpectedSpaceState{
			0: {1, 0, 0},
			1: {1, 3, 0},
			4: {1, -2, 0},
			5: {1, -10, 0},
			6: {1, -19, 0},
			7: {1, -25, 0},

			8:  {1, -51 + deltaT8FromIdx0, 0},
			9:  {1, -69 + deltaT9FromIdx0, 0},
			10: {1, -79 + deltaT10FromIdx0, 0},
			11: {1, -131 + deltaT11FromIdx0, 0},
		}
		assertSpaceStates(t, &space, expectedState, 10, getContextString(evt.growthContext))

		assertNearMainPoints(t, &space)
	}
}

func getContextString(ctx GrowthContext) string {
	return fmt.Sprintf("Type %d, Idx %d", ctx.permutationType, ctx.permutationIndex)
}

func TestSingleSimpleContextD0(t *testing.T) {
	Log.Level = m3util.INFO
	InitConnectionDetails()
	for trioIdx := 0; trioIdx < 8; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))
		ctx := GrowthContext{&Origin, 1, trioIdx, false, 0}
		space.CreateEventWithGrowthContext(Origin, RedEvent, ctx)

		expectedState := map[TickTime]ExpectedSpaceState{
			0:  {1, 0, 0},
			1:  {1, 3, 0},
			4:  {1, 0, 0},
			5:  {1, -13, 0},
			6:  {1, -22, 0},
			7:  {1, -27, 0},
			8:  {1, -52, 0},
			9:  {1, -64, 0},
			10: {1, -78, 0},
			11: {1, -115, 0},
			12: {1, -130, 0},
			13: {1, -153, 0},
			14: {1, -202, 0},
		}
		assertSpaceStates(t, &space, expectedState, 14, getContextString(ctx))

		assertNearMainPoints(t, &space)
	}
}

func TestSingleRedEventD1(t *testing.T) {
	Log.Level = m3util.INFO
	InitConnectionDetails()

	space := MakeSpace(3 * 9)

	assertEmptySpace(t, &space, 3*9)

	space.SetEventOutgrowthThreshold(Distance(1))

	evt := space.CreateSingleEventCenter()

	expectedState := map[TickTime]ExpectedSpaceState{
		0: {1, 0, 0},
		1: {1, 3, 0},
		2: {1, 6, 3},
		3: {1, 12, 6},
		4: {1, -2, 12},
		5: {1, -10, 22},
	}
	assertSpaceStates(t, &space, expectedState, 5, getContextString(evt.growthContext))
}

func assertEmptySpace(t *testing.T, space *Space, max int64) {
	assert.Equal(t, max, space.Max)
	assert.Equal(t, 0, len(space.activeNodesMap))
	assert.Equal(t, 0, len(space.activeConnections))
	assert.Equal(t, 0, len(space.events))
}

func assertSpaceStates(t *testing.T, space *Space, expectMap map[TickTime]ExpectedSpaceState, finalTime TickTime, contextMsg string) {
	expectedTime := TickTime(0)
	expect, ok := expectMap[expectedTime]
	assert.True(t, ok, "%s: Should have the 0 tick time map entry in %v", contextMsg, expectMap)
	baseNodes := expect.baseNodes
	newNodes := baseNodes
	activeNodes := baseNodes
	nbNodes := baseNodes
	nbConnections := 0
	nbMainPoints := baseNodes
	nbActiveMainPoints := baseNodes
	for {
		assertSpaceSingleEvent(t, space, expectedTime, nbNodes, nbConnections, activeNodes, nbMainPoints, nbActiveMainPoints, contextMsg)
		if expectedTime == finalTime {
			break
		}
		space.ForwardTime()
		expectedTime++
		if expectedTime == 3 {
			// Should reach all center face
			nbMainPoints = baseNodes + 6
			nbActiveMainPoints = baseNodes + 6
		} else if expectedTime == 4 {
			nbMainPoints = nbMainPoints + 6
			nbActiveMainPoints = baseNodes + 6
		} else if expectedTime == 5 {
			nbMainPoints = nbMainPoints + 2
			nbActiveMainPoints = baseNodes + 2
		} else if expectedTime == 6 {
			// Stop testing main points
			nbMainPoints = -1
			nbActiveMainPoints = -1
		}

		expect, ok = expectMap[expectedTime]
		if ok {
			if expect.newNodes <= 0 {
				newNodes *= 2
				newNodes += expect.newNodes
			} else {
				newNodes = expect.newNodes
			}
			activeNodes = newNodes + expect.oldActiveNodes + baseNodes
			if expectedTime == 5 && expect.newNodes == -10 {
				nbMainPoints += 2
				nbActiveMainPoints += 2
			}
			if (expectedTime == 4 || expectedTime == 5) && expect.oldActiveNodes > 0 {
				nbActiveMainPoints += 6
			}
		} else {
			newNodes *= 2
			activeNodes = newNodes + baseNodes
		}
		nbConnections += newNodes
		nbNodes += newNodes
	}
}

func assertSpaceSingleEvent(t *testing.T, space *Space, time TickTime, nbNodes, nbConnections, nbActive, nbMainPoints, nbActiveMainPoints int, contextMsg string) {
	assert.Equal(t, time, space.currentTime, contextMsg)
	assert.Equal(t, nbNodes, len(space.activeNodesMap), "%s: nbNodes failed at %d", contextMsg, time)
	assert.Equal(t, nbConnections, len(space.activeConnections), "%s: nbConnections failed at %d", contextMsg, time)
	assert.Equal(t, 1, len(space.events), "%s: nbEvents failed at %d", contextMsg, time)
	totalNodeActive := 0
	totalMainPoints := 0
	totalMainPointsActive := 0
	for _, node := range space.activeNodesMap {
		if node.IsActive(space) {
			totalNodeActive++
			// Only one color since it's single event
			assert.Equal(t, uint8(1), node.HowManyColors(space), "%s: Number of colors of node %v wrong at time %d", contextMsg, node, time)
			// The color should be red only
			assert.Equal(t, uint8(RedEvent), node.GetColorMask(space), "%s: Number of colors of node %v wrong at time %d", contextMsg, node, time)
		}
		if node.Pos.IsMainPoint() {
			totalMainPoints++
			if node.IsActive(space) {
				totalMainPointsActive++
			}
		}
	}
	assert.Equal(t, nbActive, totalNodeActive, "%s: nbActiveNodes failed at %d", contextMsg, time)
	if nbMainPoints > 0 {
		assert.Equal(t, nbMainPoints, totalMainPoints, "%s: totalMainPoints failed at %d", contextMsg, time)
		assert.Equal(t, nbActiveMainPoints, totalMainPointsActive, "%s: totalMainPointsActive failed at %d", contextMsg, time)
	}
}

func assertNearMainPoints(t *testing.T, space *Space) {
	for _, node := range space.activeNodesMap {
		// Find main Pos attached to node
		var mainPointNode *Node
		if node.Pos.IsMainPoint() {
			mainPointNode = node
		} else {
			for _, conn := range node.connections {
				if conn != nil {
					if conn.N1.Pos.IsMainPoint() {
						mainPointNode = conn.N1
						break
					}
					if conn.N2.Pos.IsMainPoint() {
						mainPointNode = conn.N2
						break
					}
				}
			}
		}
		if mainPointNode != nil {
			assert.Equal(t, node.Pos.getNearMainPoint(), *(mainPointNode.Pos))
		}
	}
}
