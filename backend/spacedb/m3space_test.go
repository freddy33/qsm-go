package spacedb

import (
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
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

func getOrCreateSpace(t *testing.T, spaceName string, threshold m3space.DistAndTime) *SpaceDb {
	env := getSpaceTestEnv()
	spaceData := GetServerSpacePackData(env)
	allSpaces := spaceData.GetAllSpaces()
	var space *SpaceDb
	for _, sp := range allSpaces {
		if sp.GetName() == spaceName {
			space = sp.(*SpaceDb)
		}
	}
	if space == nil {
		sp, err := spaceData.CreateSpace(spaceName, threshold, 4, 4)
		if !assert.NoError(t, err) {
			return nil
		}
		space = sp.(*SpaceDb)
	}
	return space
}

func Test_Evt1_Type8_D0(t *testing.T) {
	Log.SetWarn()
	LogStat.SetInfo()

	space := getOrCreateSpace(t, "Test_Evt1_Type8_D0", m3space.DistAndTime(0))
	if space == nil {
		return
	}

	for gowthIdx := 0; gowthIdx < 12; gowthIdx++ {
		evt, err := space.CreateEvent(m3point.GrowthType(8), gowthIdx, 0, m3space.DistAndTime(0), m3point.Origin, m3space.RedEvent)
		if !assert.NoError(t, err) {
			return
		}

		deltaT8FromIdx0 := 0
		deltaT9FromIdx0 := 0
		deltaT10FromIdx0 := 0
		deltaT11FromIdx0 := 0
		if space.pointData.GetAllMod8Permutations()[gowthIdx][3] != 4 {
			deltaT8FromIdx0 = 5
			deltaT9FromIdx0 = -11
			deltaT10FromIdx0 = -13
			deltaT11FromIdx0 = -22
		}

		expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
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
		// TODO: change back to 10 when #7 done
		assertSpaceStates(t, space, expectedState, 4, space.GetName()+" "+evt.String())
	}
}

/*
func Test_Evt1_Type8_D0_Old20_Same2(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()
	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(8)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := m3space.MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(m3space.DistAndTime(0))
		space.BlockOnSameEvent = 2
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = m3space.DistAndTime(20)

		evt := space.CreateEventAtZeroTime(ctxType, trioIdx, 0, m3point.Origin, m3space.RedEvent)

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
		expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
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
		// TODO: change back to 11 when #7 done
		assertSpaceStates(t, &space, expectedState, 4, evt.PathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_D0_Old20_Same3(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()

	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(8)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := m3space.MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		// Only latest counting
		space.SetEventOutgrowthThreshold(m3space.DistAndTime(0))
		space.BlockOnSameEvent = 3
		// No test of the old mechanism
		space.EventOutgrowthOldThreshold = m3space.DistAndTime(20)

		evt := space.CreateEventAtZeroTime(ctxType, trioIdx, 0, m3point.Origin, m3space.RedEvent)

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
		expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
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
		// TODO: change back to 11 when #7 done
		assertSpaceStates(t, &space, expectedState, 4, evt.PathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type1_D0_Old3_Dead9_Same4(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()
	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(1)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := m3space.MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		space.BlockOnSameEvent = 4
		// Only latest counting
		space.SetEventOutgrowthThreshold(m3space.DistAndTime(0))

		evt := space.CreateEventAtZeroTime(ctxType, trioIdx, 0, m3point.Origin, m3space.RedEvent)

		assert.Equal(t, m3space.DistAndTime(3), space.EventOutgrowthOldThreshold)

		expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
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
		// TODO: change back to 14 when #7 done
		assertSpaceStates(t, &space, expectedState, 4, evt.PathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type1_D0_Old3_Dead20_Same4(t *testing.T) {
	Log.SetWarn()
	m3space.LogStat.SetInfo()
	env := getSpaceTestEnv()

	ctxType := m3point.GrowthType(1)
	for trioIdx := 0; trioIdx < ctxType.GetNbIndexes(); trioIdx++ {
		space := m3space.MakeSpace(env, 3*9)

		assertEmptySpace(t, &space, 3*9)

		// Force to only 3
		space.MaxConnections = 3
		space.BlockOnSameEvent = 4
		// Only latest counting
		space.SetEventOutgrowthThreshold(m3space.DistAndTime(0))
		space.EventOutgrowthDeadThreshold = m3space.DistAndTime(20)

		evt := space.CreateEventAtZeroTime(ctxType, trioIdx, 0, m3point.Origin, m3space.RedEvent)

		assert.Equal(t, m3space.DistAndTime(3), space.EventOutgrowthOldThreshold)

		expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
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
		// TODO: change back to 14 when #7 done
		assertSpaceStates(t, &space, expectedState, 4, evt.PathContext.String())

		assertNearMainPoints(t, &space)
	}
}

func Test_Evt1_Type8_Idx0_D1_Old4_Same3(t *testing.T) {
	Log.SetInfo()
	env := getSpaceTestEnv()

	space := m3space.MakeSpace(env, 3*9)

	assertEmptySpace(t, &space, 3*9)

	space.SetEventOutgrowthThreshold(m3space.DistAndTime(1))
	assert.Equal(t, m3space.DistAndTime(1), space.EventOutgrowthThreshold)
	assert.Equal(t, m3space.DistAndTime(4), space.EventOutgrowthOldThreshold)

	evt := space.CreateSingleEventCenter()

	expectedState := map[m3space.DistAndTime]ExpectedSpaceState{
		0: simpleState(0, 0),
		1: simpleState(3, 0),
		2: oldActiveState(6, 3, 0),
		3: oldActiveState(12, 6, 6),
		4: oldActiveState(-2, 12, 6),
		5: oldActiveState(-10, 22, 4),
	}
	// TODO: change back to 5 when #7 done
	assertSpaceStates(t, &space, expectedState, 4, evt.PathContext.String())
}
*/

func assertSpaceStates(t *testing.T, space *SpaceDb, expectMap map[m3space.DistAndTime]ExpectedSpaceState, finalTime m3space.DistAndTime, contextMsg string) {
	expectedTime := m3space.DistAndTime(0)
	expect, ok := expectMap[expectedTime]
	if !assert.True(t, ok, "%s: Should have the 0 tick time map entry in %v", contextMsg, expectMap) {
		return
	}
	baseNodes := expect.baseNodes
	newNodes := baseNodes
	activeNodes := baseNodes
	nbNodes := baseNodes
	nbConnections := 0
	nbMainPoints := baseNodes
	nbActiveMainPoints := nbMainPoints
	for {
		spaceTime := space.GetSpaceTimeAt(expectedTime).(*SpaceTime)
		if !assertSpaceSingleEvent(t, spaceTime, expectedTime, nbNodes, nbConnections, activeNodes, nbMainPoints, nbActiveMainPoints, contextMsg) {
			return
		}
		if expectedTime == finalTime {
			break
		}
		fr := spaceTime.GetRuleAnalyzer()
		if !assert.NotNil(t, fr) {
			return
		}
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
	time                                                                m3space.DistAndTime
	totalRoots, totalNodeActive, totalMainPoints, totalMainPointsActive int
}

func (t *TestSpaceVisitor) VisitNode(node m3space.SpaceTimeNodeIfc) {
	if node.HasRoot() {
		t.totalRoots++
	}

	t.totalNodeActive++
	// Only one color since it's single event
	assert.Equal(t.t, uint8(1), node.HowManyColors(), "%s: Number of colors of node %v wrong at time %d", t.contextMsg, node, t.time)
	// The color should be red only
	assert.Equal(t.t, uint8(m3space.RedEvent), node.GetColorMask(), "%s: Number of colors of node %v wrong at time %d", t.contextMsg, node, t.time)

	point, err := node.GetPoint()
	if err != nil {
		Log.Error(err)
	}
	if point != nil && point.IsMainPoint() {
		t.totalMainPoints++
		t.totalMainPointsActive++
	}
}

func (t *TestSpaceVisitor) VisitLink(node m3space.SpaceTimeNodeIfc, srcPoint m3point.Point, connId m3point.ConnectionId) {
}

func assertSpaceSingleEvent(t *testing.T, spaceTime *SpaceTime, time m3space.DistAndTime, nbNodes, nbConnections, nbActive, nbMainPoints, nbActiveMainPoints int, contextMsg string) bool {
	good := assert.Equal(t, time, spaceTime.currentTime, contextMsg) &&
		assert.Equal(t, nbNodes, spaceTime.GetNbActiveNodes(), "%s: nbNodes failed at %d", contextMsg, time) &&
		// TODO: Change all test to use real active links when both sides are active
		//assert.Equal(t, nbConnections, space.GetNbActiveLinks(), "%s: nbConnections failed at %d", contextMsg, time)
		assert.Equal(t, 1, len(spaceTime.GetActiveEvents()), "%s: nbEvents failed at %d", contextMsg, time)
	if !good {
		return false
	}
	tv := new(TestSpaceVisitor)
	tv.t = t
	tv.contextMsg = contextMsg
	tv.time = time
	spaceTime.VisitAll(tv)

	good = assert.Equal(t, 1, tv.totalRoots, "%s: nb roots failed at %d", contextMsg, time) &&
		assert.Equal(t, nbActive, tv.totalNodeActive, "%s: nbActiveNodes failed at %d", contextMsg, time)
	if !good {
		return false
	}
	if nbMainPoints > 0 {
		good = assert.Equal(t, nbMainPoints, tv.totalMainPoints, "%s: totalMainPoints failed at %d", contextMsg, time) &&
			assert.Equal(t, nbActiveMainPoints, tv.totalMainPointsActive, "%s: totalMainPointsActive failed at %d", contextMsg, time)
		if !good {
			return false
		}
	}
	return true
}
