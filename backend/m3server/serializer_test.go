package m3server

import (
	"encoding/json"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/urlquery"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func assertSamePointMsg(t *testing.T, pMsg, respMsg *m3api.PointMsg) bool {
	return assert.Equal(t, pMsg.X, respMsg.X, "fail on X") &&
		assert.Equal(t, pMsg.Y, respMsg.Y, "fail on Y") &&
		assert.Equal(t, pMsg.Z, respMsg.Z, "fail on Z")
}

func TestPointMarshall(t *testing.T) {
	execTest(t, &serialTest{
		pMsg:       &m3api.PointMsg{X: 1, Y: -2, Z: 3},
		newRespMsg: func() proto.Message { return &m3api.PointMsg{} },
		assertEqual: func(t *testing.T, pMsg, respMsg proto.Message) bool {
			return assertSamePointMsg(t, pMsg.(*m3api.PointMsg), respMsg.(*m3api.PointMsg))
		},
		jsonValue:     "{\"x\":1,\"y\":-2,\"z\":3}",
		urlQueryValue: "x=1&y=-2&z=3",
	})
}

func assertSameSpaceMsg(t *testing.T, pMsg, respMsg *m3api.SpaceMsg) bool {
	good := assert.Equal(t, pMsg.SpaceId, respMsg.SpaceId, "fail on SpaceId") &&
		assert.Equal(t, pMsg.SpaceName, respMsg.SpaceName, "fail on SpaceName") &&
		assert.Equal(t, pMsg.ActiveThreshold, respMsg.ActiveThreshold, "fail on ActiveThreshold") &&
		assert.Equal(t, pMsg.MaxTriosPerPoint, respMsg.MaxTriosPerPoint, "fail on MaxTriosPerPoint") &&
		assert.Equal(t, pMsg.MaxNodesPerPoint, respMsg.MaxNodesPerPoint, "fail on MaxNodesPerPoint") &&
		assert.Equal(t, pMsg.MaxTime, respMsg.MaxTime, "fail on MaxTime") &&
		assert.Equal(t, pMsg.MaxCoord, respMsg.MaxCoord, "fail on MaxCoord") &&
		assert.Equal(t, len(pMsg.EventIds), len(respMsg.EventIds), "fail on len EventIds")
	if !good {
		return false
	}
	if pMsg.EventIds == nil {
		if !assert.Equal(t, []int32(nil), respMsg.EventIds, "collection of events should be nil") {
			return false
		}
	}
	for i, evtId := range pMsg.EventIds {
		if !assert.Equal(t, evtId, respMsg.EventIds[i], "fail on evet id %d", i) {
			return false
		}
	}
	return true
}

func TestArrayMarshall(t *testing.T) {
	execTest(t, &serialTest{
		pMsg: &m3api.SpaceMsg{
			SpaceId:          int32(3),
			SpaceName:        "test_ser",
			ActiveThreshold:  int32(3),
			MaxTriosPerPoint: int32(2),
			MaxNodesPerPoint: int32(5),
			MaxTime:          int32(445),
			MaxCoord:         int32(4566),
			EventIds:         []int32{5, 6, 7, 8, 9},
		},
		newRespMsg: func() proto.Message {
			return &m3api.SpaceMsg{}
		},
		assertEqual: func(t *testing.T, pMsg, respMsg proto.Message) bool {
			return assertSameSpaceMsg(t, pMsg.(*m3api.SpaceMsg), respMsg.(*m3api.SpaceMsg))
		},
		jsonValue:     "{\"space_id\":3,\"space_name\":\"test_ser\",\"active_threshold\":3,\"max_trios_per_point\":2,\"max_nodes_per_point\":5,\"max_time\":445,\"max_coord\":4566,\"event_ids\":[5,6,7,8,9]}",
		urlQueryValue: "space_id=3&space_name=test_ser&active_threshold=3&max_trios_per_point=2&max_nodes_per_point=5&max_time=445&max_coord=4566&event_ids%5B%5D=5&event_ids%5B%5D=6&event_ids%5B%5D=7&event_ids%5B%5D=8&event_ids%5B%5D=9",
	})
}

func TestEmptyArrayMarshall(t *testing.T) {
	execTest(t, &serialTest{
		pMsg: &m3api.SpaceMsg{
			SpaceId:         int32(4),
			SpaceName:       "test_empty_ser",
			ActiveThreshold: int32(2),
			EventIds:        nil,
		},
		newRespMsg: func() proto.Message {
			return &m3api.SpaceMsg{}
		},
		assertEqual: func(t *testing.T, pMsg, respMsg proto.Message) bool {
			return assertSameSpaceMsg(t, pMsg.(*m3api.SpaceMsg), respMsg.(*m3api.SpaceMsg))
		},
		jsonValue:     "{\"space_id\":4,\"space_name\":\"test_empty_ser\",\"active_threshold\":2,\"max_trios_per_point\":0,\"max_nodes_per_point\":0,\"max_time\":0,\"max_coord\":0}",
		urlQueryValue: "space_id=4&space_name=test_empty_ser&active_threshold=2",
	})
}

func assertSameCreateEventMsg(t *testing.T, pMsg, respMsg *m3api.CreateEventRequestMsg) bool {
	good := assert.Equal(t, pMsg.SpaceId, respMsg.SpaceId, "fail on SpaceId") &&
		assert.Equal(t, pMsg.GrowthType, respMsg.GrowthType, "fail on GrowthType") &&
		assert.Equal(t, pMsg.GrowthOffset, respMsg.GrowthOffset, "fail on GrowthOffset") &&
		assert.Equal(t, pMsg.GrowthIndex, respMsg.GrowthIndex, "fail on GrowthIndex") &&
		assert.Equal(t, pMsg.CreationTime, respMsg.CreationTime, "fail on CreationTime") &&
		assert.Equal(t, pMsg.Color, respMsg.Color, "fail on Color")
	if !good {
		return false
	}
	if pMsg.Center == nil {
		if !assert.Equal(t, (*m3api.PointMsg)(nil), respMsg.Center, "center point should be nil") {
			return false
		}
		return true
	}
	return good &&
		assert.Equal(t, pMsg.Center.X, respMsg.Center.X, "fail on X") &&
		assert.Equal(t, pMsg.Center.Y, respMsg.Center.Y, "fail on Y") &&
		assert.Equal(t, pMsg.Center.Z, respMsg.Center.Z, "fail on Z")
}

func TestComplexMarshall(t *testing.T) {
	execTest(t, &serialTest{
		pMsg: &m3api.CreateEventRequestMsg{
			SpaceId:      int32(22),
			GrowthType:   int32(8),
			GrowthIndex:  int32(8),
			GrowthOffset: int32(3),
			CreationTime: int32(2),
			Center: &m3api.PointMsg{
				X: 1,
				Y: -2,
				Z: 3,
			},
			Color: uint32(2),
		},
		newRespMsg: func() proto.Message {
			return &m3api.CreateEventRequestMsg{}
		},
		assertEqual: func(t *testing.T, pMsg, respMsg proto.Message) bool {
			return assertSameCreateEventMsg(t, pMsg.(*m3api.CreateEventRequestMsg), respMsg.(*m3api.CreateEventRequestMsg))
		},
		jsonValue:     "{\"space_id\":22,\"growth_type\":8,\"growth_index\":8,\"growth_offset\":3,\"creation_time\":2,\"center\":{\"x\":1,\"y\":-2,\"z\":3},\"color\":2}",
		urlQueryValue: "space_id=22&growth_type=8&growth_index=8&growth_offset=3&creation_time=2&center%5Bx%5D=1&center%5By%5D=-2&center%5Bz%5D=3&color=2",
	})
}

func TestEmptyComplexMarshall(t *testing.T) {
	execTest(t, &serialTest{
		pMsg: &m3api.CreateEventRequestMsg{
			SpaceId:    int32(23),
			GrowthType: int32(2),
			Center:     nil,
			Color:      uint32(1),
		},
		newRespMsg: func() proto.Message {
			return &m3api.CreateEventRequestMsg{}
		},
		assertEqual: func(t *testing.T, pMsg, respMsg proto.Message) bool {
			return assertSameCreateEventMsg(t, pMsg.(*m3api.CreateEventRequestMsg), respMsg.(*m3api.CreateEventRequestMsg))
		},
		jsonValue:     "{\"space_id\":23,\"growth_type\":2,\"growth_index\":0,\"growth_offset\":0,\"creation_time\":0,\"color\":1}",
		urlQueryValue: "space_id=23&growth_type=2&color=1",
	})
}

type serialTest struct {
	pMsg          proto.Message
	newRespMsg    func() proto.Message
	assertEqual   func(t *testing.T, pMsg, respMsg proto.Message) bool
	jsonValue     string
	urlQueryValue string
}

func execTest(t *testing.T, test *serialTest) bool {
	// Protobuf test
	b, err := proto.Marshal(test.pMsg)
	good := assert.NoError(t, err) &&
		assert.True(t, len(b) > 3)
	if !good {
		return false
	}
	respMsg := test.newRespMsg()
	err = proto.Unmarshal(b, respMsg)
	good = assert.NoError(t, err) &&
		test.assertEqual(t, test.pMsg, respMsg)
	if !good {
		return false
	}

	// JSON Test
	b, err = json.Marshal(test.pMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, test.jsonValue, string(b))
	if !good {
		return false
	}
	respMsg = test.newRespMsg()
	err = json.Unmarshal(b, respMsg)
	good = assert.NoError(t, err) &&
		test.assertEqual(t, test.pMsg, respMsg)
	if !good {
		return false
	}

	b, err = urlquery.Marshal(test.pMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, test.urlQueryValue, string(b))
	if !good {
		return false
	}

	respMsg = test.newRespMsg()
	err = urlquery.Unmarshal(b, respMsg)
	good = assert.NoError(t, err) &&
		test.assertEqual(t, test.pMsg, respMsg)
	if !good {
		return false
	}

	return true
}
