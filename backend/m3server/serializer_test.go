package m3server

import (
	"encoding/json"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/golang/protobuf/proto"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPointMarshall(t *testing.T) {
	pMsg := &m3api.PointMsg{
		X: 1,
		Y: -2,
		Z: 3,
	}
	b, err := proto.Marshal(pMsg)
	good := assert.NoError(t, err) &&
		assert.True(t, len(b) > 3)
	if !good {
		return
	}
	respMsg := &m3api.PointMsg{}
	err = proto.Unmarshal(b, respMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, pMsg.X, respMsg.X, "fail on X") &&
		assert.Equal(t, pMsg.Y, respMsg.Y, "fail on Y") &&
		assert.Equal(t, pMsg.Z, respMsg.Z, "fail on Z")

	b, err = json.Marshal(pMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, "{\"x\":1,\"y\":-2,\"z\":3}", string(b))
	if !good {
		return
	}
	respMsg = &m3api.PointMsg{}
	err = json.Unmarshal(b, respMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, pMsg.X, respMsg.X, "fail on X") &&
		assert.Equal(t, pMsg.Y, respMsg.Y, "fail on Y") &&
		assert.Equal(t, pMsg.Z, respMsg.Z, "fail on Z")

	b, err = urlquery.Marshal(pMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, "x=1&y=-2&z=3", string(b))
	if !good {
		return
	}
	respMsg = &m3api.PointMsg{}
	err = urlquery.Unmarshal(b, respMsg)
	good = assert.NoError(t, err) &&
		assert.Equal(t, pMsg.X, respMsg.X, "fail on X") &&
		assert.Equal(t, pMsg.Y, respMsg.Y, "fail on Y") &&
		assert.Equal(t, pMsg.Z, respMsg.Z, "fail on Z")
}
