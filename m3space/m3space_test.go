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
	deadNodes      int
	mainPoints     int
}

func simpleState(newNodes int, mainPoints int) ExpectedSpaceState {
	return ExpectedSpaceState{1, newNodes, 0, 0, mainPoints}
}

func noMainState(newNodes int) ExpectedSpaceState {
	return ExpectedSpaceState{1, newNodes, 0, 0, -1}
}

func oldActiveState(newNodes int, oldActiveNodes int, mainPoints int) ExpectedSpaceState {
	return ExpectedSpaceState{1, newNodes, oldActiveNodes, 0, mainPoints}
}

func deadState(newNodes int, deadNodes int) ExpectedSpaceState {
	return ExpectedSpaceState{1, newNodes, 0, deadNodes, -1}
}

func Test_Evt1_Type8_D0_Old20_Same4(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	for trioIdx := 0; trioIdx < 12; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))
		space.blockOnSameEvent = 4
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = Distance(20)

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
			0: simpleState(0, 0),
			1: simpleState(3, 0),
			3: simpleState(0, 6),
			4: simpleState(-2, 6),
			5: simpleState(-10, 4),
			6: noMainState(-19),
			7: noMainState(-25),

			8:  noMainState(-51 + deltaT8FromIdx0),
			9:  noMainState(-69 + deltaT9FromIdx0),
			10: noMainState(-79 + deltaT10FromIdx0),
			11: noMainState(-131 + deltaT11FromIdx0),
		}
		assertSpaceStates(t, &space, expectedState, 10, getContextString(evt.growthContext))

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_D0_Old20_Same2(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	for trioIdx := 0; trioIdx < 8; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))
		space.blockOnSameEvent = 2
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = Distance(20)

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
		if AllMod8Permutations[trioIdx][5] == 5 {
			deltaT8FromIdx0 += 2
			deltaT9FromIdx0 -= 10
			deltaT10FromIdx0 += 24
			deltaT11FromIdx0 += 37
		}

		expectedState := map[TickTime]ExpectedSpaceState{
			0: simpleState(0, 0),
			1: simpleState(3, 0),
			3: simpleState(0, 6),
			4: simpleState(-4, 6),
			5: simpleState(-16, 4),
			6: noMainState(-17),
			7: noMainState(-16),

			8:  noMainState(-31 + deltaT8FromIdx0),
			9:  noMainState(-31 + deltaT9FromIdx0),
			10: noMainState(-61 + deltaT10FromIdx0),
			11: noMainState(-102 + deltaT11FromIdx0),
		}
		assertSpaceStates(t, &space, expectedState, 11, getContextString(evt.growthContext))

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_D0_Old20_Same3(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	for trioIdx := 0; trioIdx < 8; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))
		space.blockOnSameEvent = 3
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = Distance(20)

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
		if AllMod8Permutations[trioIdx][5] == 5 {
			deltaT9FromIdx0 -= 4
			deltaT10FromIdx0 += 11
			deltaT11FromIdx0 += 44
		}

		expectedState := map[TickTime]ExpectedSpaceState{
			0: simpleState(0, 0),
			1: simpleState(3, 0),
			3: simpleState(0, 6),
			4: simpleState(-2, 6),
			5: simpleState(-10, 4),
			6: noMainState(-20),
			7: noMainState(-23),

			8:  noMainState(-54 + deltaT8FromIdx0),
			9:  noMainState(-65 + deltaT9FromIdx0),
			10: noMainState(-80 + deltaT10FromIdx0),
			11: noMainState(-133 + deltaT11FromIdx0),
		}
		assertSpaceStates(t, &space, expectedState, 11, getContextString(evt.growthContext))

		assertNearMainPoints(t, &space)
	}
}

func getContextString(ctx *GrowthContext) string {
	return fmt.Sprintf("Type %d, Idx %d", ctx.permutationType, ctx.permutationIndex)
}

func Test_Evt1_Type1_D0_Old3_Dead9_Same4(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	for trioIdx := 0; trioIdx < 8; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		space.blockOnSameEvent = 4
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))

		ctx := GrowthContext{Origin, 1, trioIdx, false, 0}
		space.CreateEventWithGrowthContext(Origin, RedEvent, &ctx)

		assert.Equal(t, Distance(3), space.EventOutgrowthOldThreshold)

		expectedState := map[TickTime]ExpectedSpaceState{
			0:  simpleState(0, 0),
			1:  simpleState(3, 0),
			3:  simpleState(0, 6),
			4:  simpleState(0, 6),
			5:  simpleState(-13, 2),
			6:  simpleState(-22, -1),
			7:  simpleState(-27, -1),
			8:  simpleState(-52, -1),
			9:  simpleState(-64, -1),
			10: deadState(-78, 3),
			11: deadState(-115, 6),
			12: deadState(-130, 12),
			13: deadState(-153, 24),
			14: deadState(-202, 35),
		}
		assertSpaceStates(t, &space, expectedState, 14, getContextString(&ctx))

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type1_D0_Old3_Dead20_Same4(t *testing.T) {
	Log.Level = m3util.WARN
	LogStat.Level = m3util.INFO
	for trioIdx := 0; trioIdx < 8; trioIdx++ {
		space := MakeSpace(3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		space.blockOnSameEvent = 4
		// Only latest counting
		space.SetEventOutgrowthThreshold(Distance(0))
		space.EventOutgrowthDeadThreshold = Distance(20)

		ctx := GrowthContext{Origin, 1, trioIdx, false, 0}
		space.CreateEventWithGrowthContext(Origin, RedEvent, &ctx)

		assert.Equal(t, Distance(3), space.EventOutgrowthOldThreshold)

		expectedState := map[TickTime]ExpectedSpaceState{
			0:  simpleState(0, 0),
			1:  simpleState(3, 0),
			3:  simpleState(0, 6),
			4:  simpleState(0, 6),
			5:  simpleState(-13, 2),
			6:  simpleState(-22, -1),
			7:  simpleState(-27, -1),
			8:  simpleState(-52, -1),
			9:  simpleState(-64, -1),
			10: simpleState(-78, -1),
			11: simpleState(-115, -1),
			12: simpleState(-130, -1),
			13: simpleState(-153, -1),
			14: simpleState(-202, -1),
		}
		assertSpaceStates(t, &space, expectedState, 14, getContextString(&ctx))

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_Idx0_D1_Old4_Same3(t *testing.T) {
	Log.Level = m3util.INFO

	space := MakeSpace(3 * 9)

	assertEmptySpace(t, &space, 3*9)

	space.SetEventOutgrowthThreshold(Distance(1))
	assert.Equal(t, Distance(1), space.EventOutgrowthThreshold)
	assert.Equal(t, Distance(4), space.EventOutgrowthOldThreshold)

	evt := space.CreateSingleEventCenter()
	assert.Equal(t, uint8(8), evt.growthContext.permutationType)
	assert.Equal(t, 0, evt.growthContext.permutationIndex)
	assert.Equal(t, 0, evt.growthContext.permutationOffset)
	assert.Equal(t, false, evt.growthContext.permutationNegFlow)
	assert.Equal(t, Origin, evt.growthContext.center)

	expectedState := map[TickTime]ExpectedSpaceState{
		0: simpleState(0, 0),
		1: simpleState(3, 0),
		2: oldActiveState(6, 3, 0),
		3: oldActiveState(12, 6, 6),
		4: oldActiveState(-2, 12, 6),
		5: oldActiveState(-10, 22, 4),
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
	nbActiveMainPoints := nbMainPoints
	for {
		assertSpaceSingleEvent(t, space, expectedTime, nbNodes, nbConnections, activeNodes, nbMainPoints, nbActiveMainPoints, contextMsg)
		if expectedTime == finalTime {
			break
		}
		space.ForwardTime()
		expectedTime++

		expect, ok = expectMap[expectedTime]
		if ok {
			if expect.newNodes <= 0 {
				newNodes *= 2
				newNodes += expect.newNodes
			} else {
				newNodes = expect.newNodes
			}
			activeNodes = newNodes + expect.oldActiveNodes + baseNodes
			if expect.mainPoints >= 0 {
				nbMainPoints += expect.mainPoints
				nbActiveMainPoints = baseNodes + expect.mainPoints
				if expect.oldActiveNodes > 0 {
					oldExpect, ok := expectMap[expectedTime-1]
					if ok {
						nbActiveMainPoints += oldExpect.mainPoints
					}
				}
			} else {
				// Stop testing main points
				nbMainPoints = -1
				nbActiveMainPoints = -1
			}
			if expect.deadNodes > 0 {
				nbNodes -= expect.deadNodes
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
	assert.Equal(t, nbNodes, space.GetNbNodes(), "%s: nbNodes failed at %d", contextMsg, time)
	assert.Equal(t, nbConnections, space.GetNbConnections(), "%s: nbConnections failed at %d", contextMsg, time)
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
		var mainPointNode *ActiveNode
		if node.Pos.IsMainPoint() {
			mainPointNode = node
		} else {
			for _, conn := range node.connections {
				bv, ok := AllConnectionsIds[conn]
				assert.True(t, ok, "Failed finding for %d", conn)
				P := node.Pos.Add(bv.Vector)
				if P.IsMainPoint() {
					mainPointNode = space.getAndActivateNode(P)
					break
				}
			}
		}
		if mainPointNode != nil {
			assert.Equal(t, node.Pos.getNearMainPoint(), mainPointNode.Pos)
		}
	}
}
