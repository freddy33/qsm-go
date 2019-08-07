package m3space

import (
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
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
	Log.SetWarn()
	LogStat.SetInfo()

	env := getSpaceTestEnv()

	for trioIdx := 0; trioIdx < 12; trioIdx++ {
		space := MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(DistAndTime(0))
		space.blockOnSameEvent = 4
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = DistAndTime(20)

		evt := space.CreateEvent(8, trioIdx, 0, m3point.Origin, RedEvent)

		deltaT8FromIdx0 := 0
		deltaT9FromIdx0 := 0
		deltaT10FromIdx0 := 0
		deltaT11FromIdx0 := 0
		if m3point.AllMod8Permutations[trioIdx][3] != 4 {
			deltaT8FromIdx0 = 5
			deltaT9FromIdx0 = -11
			deltaT10FromIdx0 = -13
			deltaT11FromIdx0 = -22
		}

		expectedState := map[DistAndTime]ExpectedSpaceState{
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
		assertSpaceStates(t, &space, expectedState, 10, evt.pathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_D0_Old20_Same2(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(8)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(DistAndTime(0))
		space.blockOnSameEvent = 2
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = DistAndTime(20)

		evt := space.CreateEvent(ctxType, trioIdx, 0, m3point.Origin, RedEvent)

		deltaT8FromIdx0 := 0
		deltaT9FromIdx0 := 0
		deltaT10FromIdx0 := 0
		deltaT11FromIdx0 := 0
		// All odd indexes have different behavior
		if trioIdx%2 == 1 {
			deltaT8FromIdx0 = 5
			deltaT9FromIdx0 = -11
			deltaT10FromIdx0 = -13
			deltaT11FromIdx0 = 22
		}
		expectedState := map[DistAndTime]ExpectedSpaceState{
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
		assertSpaceStates(t, &space, expectedState, 11, evt.pathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_D0_Old20_Same3(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()

	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(8)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(DistAndTime(0))
		space.blockOnSameEvent = 3
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = DistAndTime(20)

		evt := space.CreateEvent(ctxType, trioIdx, 0, m3point.Origin, RedEvent)

		deltaT8FromIdx0 := 0
		deltaT9FromIdx0 := 0
		deltaT10FromIdx0 := 0
		deltaT11FromIdx0 := 0
		// All odd indexes have different behavior
		if trioIdx%2 == 1 {
			deltaT8FromIdx0 = 5
			deltaT9FromIdx0 = -11
			deltaT10FromIdx0 = -13
			deltaT11FromIdx0 = 22
		}
		expectedState := map[DistAndTime]ExpectedSpaceState{
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
		assertSpaceStates(t, &space, expectedState, 11, evt.pathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type1_D0_Old3_Dead9_Same4(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(1)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		space.blockOnSameEvent = 4
		// Only latest counting
		space.SetEventOutgrowthThreshold(DistAndTime(0))

		evt := space.CreateEvent(ctxType, trioIdx, 0, m3point.Origin, RedEvent)

		assert.Equal(t, DistAndTime(3), space.EventOutgrowthOldThreshold)

		expectedState := map[DistAndTime]ExpectedSpaceState{
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
		// TODO: Manage dead path element (final time should be 14)
		assertSpaceStates(t, &space, expectedState, 9, evt.pathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type1_D0_Old3_Dead20_Same4(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()
	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(1)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := MakeSpace(env,3 * 9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		space.blockOnSameEvent = 4
		// Only latest counting
		space.SetEventOutgrowthThreshold(DistAndTime(0))
		space.EventOutgrowthDeadThreshold = DistAndTime(20)

		evt := space.CreateEvent(ctxType, trioIdx, 0, m3point.Origin, RedEvent)

		assert.Equal(t, DistAndTime(3), space.EventOutgrowthOldThreshold)

		expectedState := map[DistAndTime]ExpectedSpaceState{
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
		assertSpaceStates(t, &space, expectedState, 14, evt.pathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_Idx0_D1_Old4_Same3(t *testing.T) {
	Log.SetInfo()
	env := getSpaceTestEnv()

	space := MakeSpace(env, 3 * 9)

	assertEmptySpace(t, &space, 3*9)

	space.SetEventOutgrowthThreshold(DistAndTime(1))
	assert.Equal(t, DistAndTime(1), space.EventOutgrowthThreshold)
	assert.Equal(t, DistAndTime(4), space.EventOutgrowthOldThreshold)

	evt := space.CreateSingleEventCenter()

	expectedState := map[DistAndTime]ExpectedSpaceState{
		0: simpleState(0, 0),
		1: simpleState(3, 0),
		2: oldActiveState(6, 3, 0),
		3: oldActiveState(12, 6, 6),
		4: oldActiveState(-2, 12, 6),
		5: oldActiveState(-10, 22, 4),
	}
	assertSpaceStates(t, &space, expectedState, 5, evt.pathContext.String())
}

func assertEmptySpace(t *testing.T, space *Space, max m3point.CInt) {
	assert.Equal(t, max, space.Max)
	assert.Equal(t, 0, len(space.activeNodes))
	assert.Equal(t, 0, len(space.activeLinks))
	assert.Equal(t, 0, space.GetNbEvents())
}

func assertSpaceStates(t *testing.T, space *Space, expectMap map[DistAndTime]ExpectedSpaceState, finalTime DistAndTime, contextMsg string) {
	expectedTime := DistAndTime(0)
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

type TestSpaceVisitor struct {
	t                                                                   *testing.T
	contextMsg                                                          string
	time                                                                DistAndTime
	totalRoots, totalNodeActive, totalMainPoints, totalMainPointsActive int
}

func (t *TestSpaceVisitor) VisitNode(space *Space, node Node) {
	if node.HasRoot() {
		t.totalRoots++
	}
	if node.IsActive(space) {
		t.totalNodeActive++
		// Only one color since it's single event
		assert.Equal(t.t, uint8(1), node.HowManyColors(space), "%s: Number of colors of node %v wrong at time %d", t.contextMsg, node, t.time)
		// The color should be red only
		assert.Equal(t.t, uint8(RedEvent), node.GetColorMask(space), "%s: Number of colors of node %v wrong at time %d", t.contextMsg, node, t.time)
	}
	if node.GetPoint().IsMainPoint() {
		t.totalMainPoints++
		if node.IsActive(space) {
			t.totalMainPointsActive++
		}
	}
}

func (t *TestSpaceVisitor) VisitLink(space *Space, pl m3path.PathLink) {
}

func assertSpaceSingleEvent(t *testing.T, space *Space, time DistAndTime, nbNodes, nbConnections, nbActive, nbMainPoints, nbActiveMainPoints int, contextMsg string) {
	assert.Equal(t, time, space.currentTime, contextMsg)
	assert.Equal(t, nbNodes, space.GetNbNodes(), "%s: nbNodes failed at %d", contextMsg, time)
	// TODO: Change all test to use real active links when both sides are active
	//assert.Equal(t, nbConnections, space.GetNbActiveLinks(), "%s: nbConnections failed at %d", contextMsg, time)
	assert.Equal(t, 1, space.GetNbEvents(), "%s: nbEvents failed at %d", contextMsg, time)
	tv := new(TestSpaceVisitor)
	tv.t = t
	tv.contextMsg = contextMsg
	tv.time = time
	space.VisitAll(tv, false)

	assert.Equal(t, 1, tv.totalRoots, "%s: nb roots failed at %d", contextMsg, time)
	assert.Equal(t, nbActive, tv.totalNodeActive, "%s: nbActiveNodes failed at %d", contextMsg, time)
	if nbMainPoints > 0 {
		assert.Equal(t, nbMainPoints, tv.totalMainPoints, "%s: totalMainPoints failed at %d", contextMsg, time)
		assert.Equal(t, nbActiveMainPoints, tv.totalMainPointsActive, "%s: totalMainPointsActive failed at %d", contextMsg, time)
	}
}

func assertNearMainPoints(t *testing.T, space *Space) {
	//nothing to test here
}
