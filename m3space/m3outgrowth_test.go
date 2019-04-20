package m3space

import (
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActiveEventOutgrowth(t *testing.T) {
	Log.Level = m3util.TRACE
	var o Outgrowth
	aeo := MakeActiveOutgrowth(m3point.Point{1, 2, 3}, Distance(0), EventOutgrowthLatest)
	o = aeo

	assert.Equal(t, m3point.Point{1, 2, 3}, o.GetPoint())
	assert.Equal(t, Distance(0), o.GetDistance())
	assert.Equal(t, EventOutgrowthLatest, o.GetState())
	assert.Equal(t, true, o.IsRoot())

	assert.Equal(t, Distance(0), o.DistanceFromLatest(nil))
	assert.Equal(t, true, o.IsActive(nil))
	assert.Equal(t, false, o.IsOld(nil))

	assert.Equal(t, false, o.HasFrom())
	assert.Equal(t, 0, o.FromLength())

	assert.Equal(t, 0, len(o.GetFromConnIds()))
	assert.Equal(t, false, o.CameFromPoint(m3point.Origin))

	o.AddFrom(m3point.Point{2, 2, 3})
	aeo.rootPath = nil

	assert.Equal(t, m3point.Point{1, 2, 3}, o.GetPoint())
	assert.Equal(t, Distance(0), o.GetDistance())
	assert.Equal(t, EventOutgrowthLatest, o.GetState())
	assert.Equal(t, false, o.IsRoot())

	assert.Equal(t, true, o.HasFrom())
	assert.Equal(t, 1, o.FromLength())

	assert.Equal(t, 1, len(o.GetFromConnIds()))
	assert.Equal(t, false, o.CameFromPoint(m3point.Origin))
	assert.Equal(t, true, o.CameFromPoint(m3point.Point{2, 2, 3}))
}

func TestSavedEventOutgrowth(t *testing.T) {
	Log.Level = m3util.TRACE
	var o Outgrowth
	o = &SavedEventOutgrowth{m3point.Point{1, 2, 3}, nil, Distance(0), m3path.TheEnd}

	assert.Equal(t, m3point.Point{1, 2, 3}, o.GetPoint())
	assert.Equal(t, Distance(0), o.GetDistance())
	assert.Equal(t, EventOutgrowthOld, o.GetState())
	assert.Equal(t, true, o.IsRoot())

	//assert.Equal(t, Distance(0), o.DistanceFromLatest(nil))
	assert.Equal(t, true, o.IsActive(nil))
	assert.Equal(t, false, o.IsOld(nil))

	assert.Equal(t, false, o.HasFrom())
	assert.Equal(t, 0, o.FromLength())

	assert.Equal(t, 0, len(o.GetFromConnIds()))
	assert.Equal(t, false, o.CameFromPoint(m3point.Origin))
}

func TestActiveEventOutgrowthPath(t *testing.T) {
	Log.Level = m3util.TRACE
	space := MakeSpace(3 * 9)
	assertEmptySpace(t, &space, 3*9)
	space.SetEventOutgrowthThreshold(Distance(0))
	//space.blockOnSameEvent = 4
	assert.Equal(t, Distance(0), space.EventOutgrowthThreshold)
	assert.Equal(t, Distance(3), space.EventOutgrowthOldThreshold)
	assert.Equal(t, 3, space.blockOnSameEvent)

	// Test center is overridden
	ctx := m3path.CreateGrowthContext(m3point.Point{5, 6, 7}, 1, 0, 0)
	evt := space.CreateEventWithGrowthContext(m3point.Origin, RedEvent, ctx)

	assert.Equal(t, m3point.Origin, ctx.GetCenter())

	assert.Equal(t, 1, len(evt.latestOutgrowths))
	assert.Equal(t, 0, len(evt.currentOutgrowths))
	assert.Equal(t, TickTime(0), evt.created)
	eo := evt.latestOutgrowths[0]

	var o Outgrowth
	o = eo

	assert.Equal(t, m3point.Point{0, 0, 0}, o.GetPoint())
	assert.Equal(t, Distance(0), o.GetDistance())
	assert.Equal(t, EventOutgrowthLatest, o.GetState())
	assert.Equal(t, true, o.IsRoot())

	assert.Equal(t, Distance(0), o.DistanceFromLatest(evt))
	assert.Equal(t, true, o.IsActive(evt))
	assert.Equal(t, false, o.IsOld(evt))

	assert.Equal(t, false, o.HasFrom())
	assert.Equal(t, 0, o.FromLength())

	assert.Equal(t, 0, len(o.GetFromConnIds()))
	assert.Equal(t, false, o.CameFromPoint(m3point.Origin))

	nextPoint := m3point.Point{1, 1, 0}
	nextPoints := evt.growthContext.GetNextPoints(o.GetPoint())
	assert.Equal(t, 3, len(nextPoints))
	assert.Equal(t, nextPoint, nextPoints[0])
	assert.Equal(t, false, o.CameFromPoint(nextPoint))

	c := make(chan *NewPossibleOutgrowth, 10)
	c <- &NewPossibleOutgrowth{nextPoint, evt, eo, eo.distance + 1, EventOutgrowthNew}
	c <- &NewPossibleOutgrowth{evt.node.Pos, evt, nil, Distance(0), EventOutgrowthEnd}

	collector := space.processNewOutgrowth(c, 1)

	assert.Equal(t, 1, len(collector.single.data))
	assert.Equal(t, 0, len(collector.sameEvent.data))
	assert.Equal(t, 0, len(collector.multiEvents.data))

	space.currentTime++

	space.realizeAllOutgrowth(collector)

	assert.Equal(t, 1, collector.single.originalPoints)
	assert.Equal(t, 0, collector.single.occupiedPoints)
	assert.Equal(t, 0, collector.single.noMoreConnPoints)
	assert.Equal(t, 0, collector.sameEvent.originalPoints)
	assert.Equal(t, 0, collector.multiEvents.originalPoints)

	// Switch latest to old, and new to latest
	for _, evt := range space.events {
		evt.moveNewOutgrowthsToLatest()
	}
	//space.moveOldToOldMaps()

	assert.Equal(t, 1, len(evt.latestOutgrowths))
	assert.Equal(t, 1, len(evt.currentOutgrowths))

	eo1 := *evt.latestOutgrowths[0]
	var o1 Outgrowth
	o1 = &eo1

	assert.Equal(t, nextPoint, o1.GetPoint())
	assert.Equal(t, Distance(1), o1.GetDistance())
	assert.Equal(t, EventOutgrowthLatest, o1.GetState())
	assert.Equal(t, EventOutgrowthCurrent, o.GetState())
	assert.Equal(t, false, o1.IsRoot())
	assert.Equal(t, true, o.IsRoot())

	assert.Equal(t, Distance(0), o1.DistanceFromLatest(evt))
	assert.Equal(t, Distance(1), o.DistanceFromLatest(evt))
	assert.Equal(t, true, o1.IsActive(evt))
	assert.Equal(t, false, o1.IsOld(evt))

	assert.Equal(t, true, o1.HasFrom())
	assert.Equal(t, 1, o1.FromLength())

	ids := o1.GetFromConnIds()
	LogTest.Infof("from conn list %v", ids)
	// TODO: All the following assertions very shaky depending on previous tests ?!?!
	/*
		assert.Equal(t, 1, len(ids))
		assert.Equal(t, int8(4), ids[0])
		assert.Equal(t, true, o1.CameFromPoint(m3point.Origin))
		p := o1.GetRootPathElement(evt)
		assert.Equal(t, 0, p.GetLength())
		assert.Equal(t, 1, p.NbForwardElements())
		assert.Equal(t, int8(-4), p.GetForwardConnId(0))
	*/
}

func TestOverlapSameEvent(t *testing.T) {
	LogStat.Level = m3util.WARN
	Log.Level = m3util.TRACE
	space := MakeSpace(3 * 9)

	assertEmptySpace(t, &space, 3*9)

	// Only latest counting
	space.SetEventOutgrowthThreshold(Distance(0))
	space.blockOnSameEvent = 4
	ctx := m3path.CreateGrowthContext(m3point.Origin, 1, 0, 0)
	space.CreateEventWithGrowthContext(m3point.Origin, RedEvent, ctx)

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
	latestOutgrowths := make(map[m3point.Point]Outgrowth, len(latestNodes))
	fromSizeHisto := make(map[int]int, 3)
	for _, evt := range space.events {
		for _, eo := range evt.latestOutgrowths {
			fromSizeHisto[eo.FromLength()]++
			lo, ok := latestOutgrowths[eo.pos]
			if ok {
				assert.Fail(t, "Should not have an outgrowth at %v for %v <= %v", eo.pos, eo.String(), lo)
			} else {
				latestOutgrowths[eo.pos] = eo
			}
		}
	}
	assert.Equal(t, nbLatestNodes-13, len(latestOutgrowths))
	Log.Info("From size histo", fromSizeHisto)

}

// Retrieve all nodes having outgrowth at exact distance d from the event
func getAllNodeWithOutgrowthAtD(space *Space, atD Distance) map[m3point.Point]Node {
	res := make(map[m3point.Point]Node, 25)
	for _, evt := range space.events {
		for _, eo := range evt.latestOutgrowths {
			if eo.distance == atD {
				res[eo.pos] = space.GetNode(eo.pos)
			}
		}
		for _, eo := range evt.currentOutgrowths {
			if eo.distance == atD {
				res[eo.pos] = space.GetNode(eo.pos)
			}
		}
	}
	return res
}
