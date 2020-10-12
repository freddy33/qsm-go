package spacedb

import (
	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var Log = m3util.NewLogger("clspace", m3util.INFO)

var envMutex sync.Mutex
var spaceEnv m3util.QsmEnvironment

func getSpaceTestEnv() m3util.QsmEnvironment {
	if spaceEnv != nil {
		return spaceEnv
	}

	envMutex.Lock()
	defer envMutex.Unlock()
	if spaceEnv != nil {
		return spaceEnv
	}
	m3util.SetToTestMode()
	spaceEnv := client.GetOrCreateInitializedApiEnv(m3util.SpaceClientTempEnv, false, true)
	return spaceEnv
}

func createNewSpace(t *testing.T, spaceName string, threshold m3space.DistAndTime) *client.SpaceCl {
	env := getSpaceTestEnv()

	spaceData := client.GetClientSpacePackData(env)
	var space *client.SpaceCl
	for _, sp := range spaceData.GetAllSpaces() {
		if sp.GetName() == spaceName {
			space = sp.(*client.SpaceCl)
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
	space = sp.(*client.SpaceCl)

	return space
}

func Test_Basic_Space(t *testing.T) {
	Log.SetDebug()
	client.Log.SetDebug()

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
	evt := event.(*client.EventCl)
	evtNode := evt.CenterNode
	if !assert.NotNil(t, evtNode, "no root node") {
		return
	}
	point, err := evtNode.GetPoint()
	td := evtNode.GetTrioDetails(client.GetClientPointPackData(space.SpaceData.Env))
	good = assert.NoError(t, err) &&
		assert.Equal(t, event.GetCenterNode().GetId(), evtNode.GetId(), "event ids do not match") &&
		assert.Equal(t, center, *point) &&
		assert.Equal(t, creationTime, evt.MaxNodeTime) &&
		assert.Equal(t, creationTime, evtNode.GetCreationTime()) &&
		assert.Equal(t, m3point.TrioIndex(1), evtNode.GetTrioIndex()) &&
		assert.Equal(t, m3point.TrioIndex(1), td.GetId())
	if !good {
		return
	}

	nodes, err := evt.GetActiveNodesAt(creationTime)
	good = assert.NoError(t, err) && assert.Equal(t, 1, len(nodes)) &&
		assert.Equal(t, evt, nodes[0].(*client.EventNodeCl).Event) &&
		assert.Equal(t, evt.CenterNode, nodes[0]) &&
		assert.Equal(t, creationTime, evt.CreationTime) &&
		assert.Equal(t, creationTime, evt.MaxNodeTime)
	if !good {
		return
	}

	time := creationTime + m3space.DistAndTime(1)
	nodes, err = evt.GetActiveNodesAt(time)
	good = assert.NoError(t, err) && assert.Equal(t, 4, len(nodes)) &&
		assert.Equal(t, evt, nodes[0].(*client.EventNodeCl).Event) &&
		// TODO: all equal except connection mask ;-)
		//assert.Equal(t, evt.CenterNode, nodes[0]) &&
		assert.Equal(t, creationTime, evt.CreationTime) &&
		assert.Equal(t, time, evt.MaxNodeTime) &&
		assert.Equal(t, m3point.CInt(22), event.GetSpace().GetMaxCoord()) &&
		assert.Equal(t, time, event.GetSpace().GetMaxTime())
	if !good {
		return
	}

	time = creationTime + m3space.DistAndTime(4)
	nodes, err = evt.GetActiveNodesAt(time)
	good = assert.NoError(t, err) && assert.Equal(t, 24-2+1, len(nodes)) &&
		assert.Equal(t, evt, nodes[0].(*client.EventNodeCl).Event) &&
		// TODO: all equal except connection mask ;-)
		//assert.Equal(t, evt.CenterNode, nodes[0]) &&
		assert.Equal(t, creationTime, evt.CreationTime) &&
		assert.Equal(t, time, evt.MaxNodeTime) &&
		assert.Equal(t, m3point.CInt(25), event.GetSpace().GetMaxCoord()) &&
		assert.Equal(t, time, event.GetSpace().GetMaxTime())
	if !good {
		return
	}
}
