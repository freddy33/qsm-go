package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
	"github.com/golang/protobuf/proto"
	"net/http"
)

var Log = m3util.NewLogger("m3server", m3util.INFO)

func GetPointPackData(env *m3db.QsmEnvironment) *m3point.PointPackData {
	if env.GetData(m3db.PointIdx) == nil {
		ppd := new(m3point.PointPackData)
		ppd.Env = env
		env.SetData(m3db.PointIdx, ppd)
		// do not return ppd but always the pointer in env data array
	}
	return env.GetData(m3db.PointIdx).(*m3point.PointPackData)
}

func GetPointData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-protobuf; messageType=backend.m3api.PointPackDataMsg")
	env := m3db.GetEnvironment(r.Context().Value(QSM_CTX_ENV_ID_KEY).(m3db.QsmEnvID))
	ppd := GetPointPackData(env)
	msg := m3api.PointPackDataMsg{}
	msg.AllConnections = make([]*m3api.ConnectionMsg, len(ppd.AllConnections))
	for idx, conn := range ppd.AllConnections {
		msg.AllConnections[idx] = &m3api.ConnectionMsg{
			ConnId: int32(conn.GetId()),
			Vector: &m3api.PointMsg{X: int32(conn.Vector.X()), Y: int32(conn.Vector.Y()), Z: int32(conn.Vector.Z())},
			Ds:     int64(conn.ConnDS),
		}
	}
	fmt.Println("sending all connections", len(msg.AllConnections))
	data, err := proto.Marshal(&msg)
	if err != nil {
		Log.Warnf("Failed to marshal Point Package Data due to: %q", err.Error())
		w.WriteHeader(500)
		_, err = fmt.Fprintf(w, "Failed to marshal Point Package Data due to:\n%s\n", err.Error())
		if err != nil {
			Log.Errorf("failed to send error message to response due to %q", err.Error())
		}
	}
	_, err = w.Write(data)
	if err != nil {
		Log.Errorf("failed to send data to response due to %q", err.Error())
	}
}
