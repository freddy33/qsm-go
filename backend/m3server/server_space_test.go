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
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestSpaceNextTime(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	qsmApp := getTestServerApp(t)
	router := qsmApp.Router

	spaceId, spaceName := callCreateSpace(t, router)
	fmt.Printf("Created %d = %q\n", spaceId, spaceName)
	if spaceId < 0 {
		return
	}

	allSpaces := callGetAllSpaces(t, router)
	if !assert.True(t, len(allSpaces) > 0) {
		return
	}
	for i, space := range allSpaces {
		Log.Infof("Index %d : Id=%d Name=%q\n", i, space.SpaceId, space.SpaceName)
	}

	eventId := callCreateEvent(t, qsmApp, spaceId, 0, m3point.Point{-3, 3, 6}, m3space.RedEvent, 8, 0, 0)
	fmt.Printf("Created %d for %d\n", eventId, spaceId)
	if eventId < 0 {
		// failed
		return
	}

	if !callDeleteSpace(t, router, spaceId, spaceName) {
		return
	}
}

func callCreateSpace(t *testing.T, router *mux.Router) (int, string) {
	rand100 := int(rand.Int31n(int32(10000)))
	if rand100 < 0 {
		rand100 = -rand100
	}
	spaceName := fmt.Sprintf("SpaceTest%02d", rand100)
	reqMsg := &m3api.SpaceMsg{
		SpaceName:        spaceName,
		ActiveThreshold:  0,
		MaxTriosPerPoint: 1,
		MaxNodesPerPoint: 4,
	}
	resMsg := &m3api.SpaceMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              router,
		requestContentType:  "proto",
		responseContentType: "proto",
		typeName:            "SpaceMsg",
		methodName:          "POST",
		uri:                 "/space",
	}, reqMsg, resMsg) {
		return -1, "failed"
	}

	spaceId := int(resMsg.SpaceId)
	good := assert.True(t, spaceId > 0, "Did not get space id id but "+strconv.Itoa(spaceId)) &&
		assert.Equal(t, spaceName, resMsg.SpaceName) &&
		assert.Equal(t, int32(0), resMsg.ActiveThreshold) &&
		assert.Equal(t, int32(1), resMsg.MaxTriosPerPoint) &&
		assert.Equal(t, int32(4), resMsg.MaxNodesPerPoint) &&
		assert.Equal(t, int32(0), resMsg.MaxTime)
	if !good {
		return -2, "failed"
	}

	return spaceId, spaceName
}

func callGetAllSpaces(t *testing.T, router *mux.Router) []*m3api.SpaceMsg {
	pMsg := &m3api.SpaceListMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              router,
		requestContentType:  "",
		responseContentType: "proto",
		typeName:            "SpaceListMsg",
		methodName:          "GET",
		uri:                 "/space",
	}, nil, pMsg) {
		return nil
	}
	return pMsg.Spaces
}

func callDeleteSpace(t *testing.T, router *mux.Router, spaceId int, spaceName string) bool {
	reqMsg := &m3api.SpaceMsg{
		SpaceId:   int32(spaceId),
		SpaceName: spaceName,
	}
	return sendAndReceive(t, &requestTest{
		router:              router,
		requestContentType:  "query",
		responseContentType: "",
		typeName:            "string",
		methodName:          "DELETE",
		uri:                 "/space",
	}, reqMsg, nil)
}

func callCreateEvent(t *testing.T, qsmApp *QsmApp, spaceId int,
	creationTime m3space.DistAndTime, point m3point.Point, color m3space.EventColor,
	growthType m3point.GrowthType, growthIndex int, growthOffset int) int {

	reqMsg := &m3api.CreateEventRequestMsg{
		SpaceId:      int32(spaceId),
		GrowthType:   int32(growthType),
		GrowthIndex:  int32(growthIndex),
		GrowthOffset: int32(growthOffset),
		CreationTime: int32(creationTime),
		Center:       m3api.PointToPointMsg(point),
		Color:        uint32(color),
	}
	resMsg := new(m3api.EventMsg)
	if !sendAndReceive(t, &requestTest{
		router:              qsmApp.Router,
		requestContentType:  "proto",
		responseContentType: "proto",
		typeName:            "EventMsg",
		methodName:          "POST",
		uri:                 "/event",
	}, reqMsg, resMsg) {
		return -1
	}
	good := assert.True(t, resMsg.EventId > 0)

	pathData := pathdb.GetServerPathPackData(qsmApp.Env)
	pathCtx := pathData.GetPathCtx(int(resMsg.GetPathCtxId()))
	good = good && assert.Equal(t, growthType, pathCtx.GetGrowthType()) &&
		assert.Equal(t, growthIndex, pathCtx.GetGrowthIndex()) &&
		assert.Equal(t, growthOffset, pathCtx.GetGrowthOffset()) &&
		assert.Equal(t, resMsg.EventId, resMsg.RootNode.EventId) &&
		assert.Equal(t, creationTime, m3space.DistAndTime(resMsg.MaxNodeTime)) &&
		assert.Equal(t, point, m3api.PointMsgToPoint(resMsg.RootNode.Point)) &&
		assert.Equal(t, int32(0), resMsg.RootNode.D)
	fmt.Println("TrioID=", resMsg.RootNode.TrioId, "ConnectionMask=", resMsg.RootNode.ConnectionMask)
	if !good {
		return -2
	}

	return int(resMsg.EventId)
}

func callNextTime(t *testing.T, spaceId int, router *mux.Router, time int, activeNodes int) bool {
	reqMsg := &m3api.SpaceTimeRequestMsg{
		SpaceId:     int32(spaceId),
		CurrentTime: int32(time),
	}
	spaceTimeResponse := &m3api.SpaceTimeResponseMsg{}
	if !sendAndReceive(t, &requestTest{
		router:              router,
		requestContentType:  "proto",
		responseContentType: "proto",
		typeName:            "SpaceTimeResponseMsg",
		methodName:          "POST",
		uri:                 "/space-time",
	}, reqMsg, spaceTimeResponse) {
		return false
	}

	return assert.Equal(t, int32(spaceId), spaceTimeResponse.GetSpaceId()) &&
		assert.Equal(t, int32(time), spaceTimeResponse.GetCurrentTime()) &&
		assert.Equal(t, activeNodes, len(spaceTimeResponse.GetActiveNodes()))
}
