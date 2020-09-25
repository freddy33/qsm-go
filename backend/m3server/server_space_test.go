package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestSpaceNextTime(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	qsmApp := getApp(m3util.SpaceTestEnv)
	router := qsmApp.Router

	initDB(t, router)

	spaceId, spaceName := callCreateSpace(t, router)
	fmt.Printf("Created %d = %q\n", spaceId, spaceName)

	allSpaces := callGetAllSpaces(t, router)
	for i, space := range allSpaces {
		fmt.Printf("Index %d : Id=%d Name=%q\n", i, space.SpaceId, space.SpaceName)
	}

	eventId := callCreateEvent(t, qsmApp, spaceId, 0, m3point.Point{-3, 3, 6}, m3space.RedEvent, 8, 0, 0)
	fmt.Printf("Created %d for %d\n", eventId, spaceId)

}

func callCreateSpace(t *testing.T, router *mux.Router) (int, string) {
	rand100 := int(rand.Int31n(int32(100)))
	if rand100 < 0 {
		rand100 = -rand100
	}
	spaceName := fmt.Sprintf("SpaceTest%02d", rand100)
	reqMsg := &m3api.SpaceMsg{
		SpaceName:               spaceName,
		ActivePathNodeThreshold: 0,
		MaxTriosPerPoint:        1,
		MaxPathNodesPerPoint:    4,
	}
	resMsg := &m3api.SpaceMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "SpaceMsg",
		methodName:  "PUT",
		uri:         "/space",
	}, reqMsg, resMsg)

	spaceId := int(resMsg.SpaceId)
	assert.True(t, spaceId > 0, "Did not get space id id but "+strconv.Itoa(spaceId))
	assert.Equal(t, spaceName, resMsg.SpaceName)
	assert.Equal(t, int32(0), resMsg.ActivePathNodeThreshold)
	assert.Equal(t, int32(1), resMsg.MaxTriosPerPoint)
	assert.Equal(t, int32(4), resMsg.MaxPathNodesPerPoint)
	assert.Equal(t, int32(0), resMsg.MaxTime)

	return spaceId, spaceName
}

func callGetAllSpaces(t *testing.T, router *mux.Router) []*m3api.SpaceMsg {
	pMsg := &m3api.SpaceListMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "SpaceListMsg",
		methodName:  "GET",
		uri:         "/space",
	}, nil, pMsg)
	assert.True(t, len(pMsg.Spaces) > 0)
	return pMsg.Spaces
}

func callDeleteSpace(t *testing.T, router *mux.Router) (int, string) {
	return -1, ""
}

func callCreateEvent(t *testing.T, qsmApp *QsmApp, spaceId int,
	time m3space.DistAndTime, point m3point.Point, color m3space.EventColor,
	growthType m3point.GrowthType, growthIndex int, growthOffset int) int {
	reqMsg := &m3api.EventMsg{
		SpaceId:      int32(spaceId),
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(growthOffset),
		CreationTime: int32(time),
		Center:       m3api.PointToPointMsg(point),
		Color:        uint32(color),
	}
	resMsg := new(m3api.EventResponseMsg)
	sendAndReceive(t, &requestTest{
		router:      qsmApp.Router,
		contentType: "proto",
		typeName:    "EventResponseMsg",
		methodName:  "PUT",
		uri:         "/event",
	}, reqMsg, resMsg)
	assert.True(t, resMsg.EventId > 0)

	pathData := pathdb.GetServerPathPackData(qsmApp.Env)
	pathCtx := pathData.GetPathCtx(int(resMsg.GetPathCtxId()))
	assert.Equal(t, growthType, pathCtx.GetGrowthType())
	assert.Equal(t, growthIndex, pathCtx.GetGrowthIndex())
	assert.Equal(t, growthOffset, pathCtx.GetGrowthOffset())

	assert.Equal(t, resMsg.EventId, resMsg.RootNode.EventId)
	assert.Equal(t, point, m3api.PointMsgToPoint(resMsg.RootNode.Point))
	assert.Equal(t, int32(0), resMsg.RootNode.D)
	fmt.Println("TrioID=", resMsg.RootNode.TrioId, "ConnectionMask=", resMsg.RootNode.ConnectionMask)

	return int(resMsg.EventId)
}

func callNextTime(t *testing.T, spaceId int, router *mux.Router, time int, activeNodes int) {
	reqMsg := &m3api.SpaceTimeRequestMsg{
		SpaceId:     int32(spaceId),
		CurrentTime: int32(time),
	}
	spaceTimeResponse := &m3api.SpaceTimeResponseMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "SpaceTimeResponseMsg",
		methodName:  "POST",
		uri:         "/space-time",
	}, reqMsg, spaceTimeResponse)

	assert.Equal(t, int32(spaceId), spaceTimeResponse.GetSpaceId())
	assert.Equal(t, int32(time), spaceTimeResponse.GetCurrentTime())
	assert.Equal(t, activeNodes, len(spaceTimeResponse.GetActiveNodes()))
}
