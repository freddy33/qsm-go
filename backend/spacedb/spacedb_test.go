package spacedb

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ExpectedSpaceState struct {
	baseNodes  int
	newNodes   int
	mainPoints int
	activeConn int
}

func simpleState(newNodes int, mainPoints int) ExpectedSpaceState {
	return ExpectedSpaceState{baseNodes: 1, newNodes: newNodes, mainPoints: mainPoints}
}

func noMainState(newNodes int) ExpectedSpaceState {
	return ExpectedSpaceState{baseNodes: 1, newNodes: newNodes, mainPoints: -1}
}

func createNewSpace(t *testing.T, spaceName string, threshold m3space.DistAndTime) *SpaceDb {
	env := getSpaceTestEnv()
	spaceData := GetServerSpacePackData(env)
	var space *SpaceDb
	for _, sp := range spaceData.GetAllSpaces() {
		if sp.GetName() == spaceName {
			space = sp.(*SpaceDb)
		}
	}

	if space != nil {
		nbDelete, err := spaceData.DeleteSpace(space.GetId(), spaceName)
		Log.Infof("Deleted %d from %s", nbDelete, space.String())
		if !assert.NoError(t, err) {
			return nil
		}
	}

	sp, err := spaceData.CreateSpace(spaceName, threshold, 4, 4)
	if !assert.NoError(t, err) {
		return nil
	}
	space = sp.(*SpaceDb)

	return space
}

func Test_Basic_Space(t *testing.T) {
	Log.SetDebug()
	pathdb.Log.SetDebug()

	space := createNewSpace(t, "Test_Basic_Space", m3space.ZeroDistAndTime)
	if space == nil {
		return
	}
	growthType := m3point.GrowthType(8)
	center := m3point.Point{6, 18, -21}
	creationTime := m3space.DistAndTime(3)
	event, err := space.CreateEvent(growthType, 3, 4, creationTime, center, m3space.BlueEvent)
	good := assert.NoError(t, err) && assert.NotNil(t, event, "event is nil") &&
		assert.Equal(t, space.String(), event.GetSpace().String()) &&
		assert.Equal(t, m3point.CInt(21), event.GetSpace().GetMaxCoord()) &&
		assert.Equal(t, creationTime, event.GetSpace().GetMaxTime()) &&
		assert.Equal(t, m3space.BlueEvent, event.GetColor()) &&
		assert.Equal(t, growthType, event.GetPathContext().GetGrowthType()) &&
		assert.Equal(t, 3, event.GetPathContext().GetGrowthIndex()) &&
		assert.Equal(t, 4, event.GetPathContext().GetGrowthOffset()) &&
		assert.NotNil(t, event.GetCenterNode(), "center node is nil")
	if !good {
		return
	}
	evt := event.(*EventDb)
	evtNode := evt.centerNode
	if !assert.NotNil(t, evtNode, "no root node") {
		return
	}
	point, err := evtNode.GetPoint()
	td := evtNode.GetTrioDetails(pointdb.GetServerPointPackData(space.spaceData.env))
	good = assert.NoError(t, err) &&
		assert.Equal(t, event.GetCenterNode().GetId(), evtNode.GetId(), "event ids do not match") &&
		assert.Equal(t, center, *point) &&
		assert.Equal(t, creationTime, evt.maxNodeTime) &&
		assert.Equal(t, creationTime, evtNode.GetCreationTime()) &&
		assert.Equal(t, m3point.TrioIndex(1), evtNode.GetTrioIndex()) &&
		assert.Equal(t, m3point.TrioIndex(1), td.GetId())
	if !good {
		return
	}

	time := creationTime
	from, to, useBetween, err := evt.getFromToTime(time)
	good = assert.NoError(t, err) && assert.False(t, useBetween, "should not be between") &&
		assert.Equal(t, TimeOnlyRoot, from, "failed on from") &&
		assert.Equal(t, TimeOnlyRoot, to, "failed on to")
	if !good {
		return
	}
	time = creationTime + m3space.DistAndTime(1)
	from, to, useBetween, err = evt.getFromToTime(time)
	good = assert.NoError(t, err) && assert.False(t, useBetween, "should not be between") &&
		assert.Equal(t, time, from, "failed on from") &&
		assert.Equal(t, time, to, "failed on to")
	if !good {
		return
	}
	time = creationTime + m3space.DistAndTime(4)
	from, to, useBetween, err = evt.getFromToTime(time)
	good = assert.NoError(t, err) && assert.False(t, useBetween, "should not be between") &&
		assert.Equal(t, time, from, "failed on from") &&
		assert.Equal(t, time, to, "failed on to")
	if !good {
		return
	}

	nodes, err := evt.GetActiveNodesDbAt(creationTime)
	good = assert.NoError(t, err) && assert.Equal(t, 1, len(nodes)) &&
		assert.Equal(t, evt, nodes[0].event) &&
		assert.Equal(t, evt.centerNode, nodes[0]) &&
		assert.Equal(t, creationTime, evt.creationTime) &&
		assert.Equal(t, creationTime, evt.maxNodeTime)
	if !good {
		return
	}

	time = creationTime + m3space.DistAndTime(1)
	nodes, err = evt.GetActiveNodesDbAt(time)
	good = assert.NoError(t, err) && assert.Equal(t, 4, len(nodes)) &&
		assert.Equal(t, evt, nodes[0].event) &&
		assert.Equal(t, evt.centerNode, nodes[0]) &&
		assert.Equal(t, creationTime, evt.creationTime) &&
		assert.Equal(t, time, evt.maxNodeTime) &&
		assert.Equal(t, m3point.CInt(22), event.GetSpace().GetMaxCoord()) &&
		assert.Equal(t, time, event.GetSpace().GetMaxTime())
	if !good {
		return
	}

	time = creationTime + m3space.DistAndTime(4)
	nodes, err = evt.GetActiveNodesDbAt(time)
	good = assert.NoError(t, err) && assert.Equal(t, 24-2+1, len(nodes)) &&
		assert.Equal(t, evt, nodes[0].event) &&
		assert.Equal(t, evt.centerNode, nodes[0]) &&
		assert.Equal(t, creationTime, evt.creationTime) &&
		assert.Equal(t, time, evt.maxNodeTime) &&
		assert.Equal(t, m3point.CInt(25), event.GetSpace().GetMaxCoord()) &&
		assert.Equal(t, time, event.GetSpace().GetMaxTime())
	if !good {
		return
	}
}

func Test_Evt1_Type8_D0(t *testing.T) {
	Log.SetWarn()
	pathdb.Log.SetWarn()
	LogStat.SetInfo()

	for growthIdx := 0; growthIdx < 12; growthIdx++ {
		space := createNewSpace(t, fmt.Sprintf("Test_TH0_Evt1_Type8_I%02d", growthIdx), m3space.DistAndTime(0))
		if space == nil {
			return
		}
		evt, err := space.CreateEvent(m3point.GrowthType(8), growthIdx, 0, m3space.ZeroDistAndTime, m3point.Origin, m3space.RedEvent)
		if !assert.NoError(t, err) {
			return
		}

		deltaT8FromIdx0 := 0
		deltaT9FromIdx0 := 0
		deltaT10FromIdx0 := 0
		deltaT11FromIdx0 := 0
		if space.pointData.GetAllMod8Permutations()[growthIdx][3] != 4 {
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
		// TODO: change back to 10 when #7 done
		if !assertSpaceStates(t, space, expectedState, 11, space.GetName()+" "+evt.String()) {
			return
		}
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

func assertInitialSpace(t *testing.T, spaceData *ServerSpacePackData, spaceName string) bool {
	var space *SpaceDb
	for _, sp := range spaceData.GetAllSpaces() {
		if sp.GetName() == spaceName {
			space = sp.(*SpaceDb)
			break
		}
	}
	if space == nil {
		return assert.Fail(t, "space not found", "Did not find any space with name %q", spaceName)
	}

	eventIds := space.GetEventIdsForMsg()
	if !assert.Equal(t, 1, len(eventIds), "fail for %s", space.String()) {
		return false
	}
	eventId := m3space.EventId(eventIds[0])
	event := space.GetEvent(eventId)
	if !assert.NotNil(t, event, "fail for space=%s eventId=%d", space.String(), eventId) {
		return false
	}
	evt := event.(*EventDb)
	msg := fmt.Sprintf("fail on space=%s event=%s", space.String(), evt.String())
	good := assert.Equal(t, m3space.ZeroDistAndTime, evt.creationTime, msg) &&
		assert.Equal(t, m3space.ZeroDistAndTime, evt.centerNode.d, msg)
	if !good {
		return false
	}
	return true
}

func assertSpaceStates(t *testing.T, space *SpaceDb, expectMap map[m3space.DistAndTime]ExpectedSpaceState, finalTime m3space.DistAndTime, contextMsg string) bool {
	expectedTime := m3space.ZeroDistAndTime
	expect, ok := expectMap[expectedTime]
	if !assert.True(t, ok, "%s: Should have the 0 tick time map entry in %v", contextMsg, expectMap) {
		return false
	}
	baseNodes := expect.baseNodes
	newNodes := baseNodes
	activeNodes := baseNodes
	nbNodes := baseNodes
	nbActiveConnections := 0
	nbMainPoints := baseNodes
	for {
		spaceTime := space.GetSpaceTimeAt(expectedTime).(*SpaceTime)
		err := spaceTime.populate()
		if !assert.NoError(t, err) {
			return false
		}
		if !assertSpaceSingleEvent(t, spaceTime, expectedTime, nbNodes, nbActiveConnections, activeNodes, nbMainPoints, contextMsg) {
			return false
		}
		if expectedTime == finalTime {
			break
		}
		fr := spaceTime.GetRuleAnalyzer()
		if !assert.NotNil(t, fr) {
			return false
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
			activeNodes = newNodes + baseNodes
			if expect.mainPoints >= 0 {
				nbMainPoints = baseNodes + expect.mainPoints
			} else {
				// Stop testing main points
				nbMainPoints = -1
			}
			nbActiveConnections += expect.activeConn
		} else {
			newNodes *= 2
			activeNodes = newNodes + baseNodes
		}
		nbNodes += newNodes
	}
	return true
}

type TestSpaceVisitor struct {
	t                                       *testing.T
	contextMsg                              string
	time                                    m3space.DistAndTime
	totalRoots, totalNodes, totalMainPoints int
	totalConn                               int
}

func (t *TestSpaceVisitor) VisitNode(node m3space.SpaceTimeNodeIfc) {
	if node.HasRoot() {
		t.totalRoots++
	}

	t.totalNodes++
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
	}
}

func (t *TestSpaceVisitor) VisitLink(node m3space.SpaceTimeNodeIfc, srcPoint m3point.Point, connId m3point.ConnectionId) {
	t.totalConn++
}

func assertSpaceSingleEvent(t *testing.T, spaceTime *SpaceTime, time m3space.DistAndTime, nbNodes, nbActiveConnections, nbActive, nbMainPoints int, contextMsg string) bool {
	totalNodes, err := spaceTime.space.GetNbNodesBetween(m3space.ZeroDistAndTime, time)
	good := assert.NoError(t, err) && assert.Equal(t, time, spaceTime.currentTime, contextMsg) &&
		assert.Equal(t, nbNodes, totalNodes, "%s nbNodes failed at %d", contextMsg, time) &&
		assert.Equal(t, nbActive, spaceTime.GetNbActiveNodes(), "%s nbActive failed at %d", contextMsg, time) &&
		assert.Equal(t, 1, len(spaceTime.GetActiveEvents()), "%s nbEvents failed at %d", contextMsg, time)
	if !good {
		return false
	}
	tv := new(TestSpaceVisitor)
	tv.t = t
	tv.contextMsg = contextMsg
	tv.time = time
	spaceTime.VisitAll(tv)

	good = assert.Equal(t, 1, tv.totalRoots, "%s nb roots failed at %d", contextMsg, time) &&
		assert.Equal(t, nbActive, tv.totalNodes, "%s nbActiveNodes failed at %d", contextMsg, time) &&
		assert.Equal(t, nbActiveConnections, tv.totalConn, "%s nbActiveConnections failed at %d", contextMsg, time)
	if !good {
		return false
	}
	if nbMainPoints > 0 {
		good = assert.Equal(t, nbMainPoints, tv.totalMainPoints, "%s totalMainPoints failed at %d", contextMsg, time)
		if !good {
			return false
		}
	}
	return true
}
