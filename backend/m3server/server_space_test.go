package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestSpaceNextTime(t *testing.T) {
	m3util.SetToTestMode()
	Log.SetInfo()
	router := getApp(m3util.SpaceTestEnv).Router

	initDB(t, router)

	spaceId, spaceName := callCreateSpace(t, router)
	fmt.Printf("Created %d = %q\n", spaceId, spaceName)
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

func callGetAllSpaces(t *testing.T, router *mux.Router) {
	pMsg := &m3api.SpaceListMsg{}
	sendAndReceive(t, &requestTest{
		router:      router,
		contentType: "proto",
		typeName:    "SpaceListMsg",
		methodName:  "GET",
		uri:         "/space",
	}, nil, pMsg)
	assert.True(t, len(pMsg.Spaces) > 0)
}

func callDeleteSpace(t *testing.T, router *mux.Router) (int, string) {
	return -1, ""
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
