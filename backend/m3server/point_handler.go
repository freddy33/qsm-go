package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/golang/protobuf/proto"
	"net/http"
)

var Log = m3util.NewLogger("m3server", m3util.INFO)

func retrievePointData(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive retrievePointData")

	w.Header().Set("Content-Type", "application/x-protobuf; messageType=model.m3api.PointPackDataMsg")

	env := GetEnvironment(r)
	ppd, _ := pointdb.GetServerPointPackData(env)
	msg := m3api.PointPackDataMsg{}

	msg.AllConnections = make([]*m3api.ConnectionMsg, len(ppd.AllConnections))
	for idx, conn := range ppd.AllConnections {
		msg.AllConnections[idx] = &m3api.ConnectionMsg{
			ConnId: int32(conn.GetId()),
			Vector: m3api.PointToPointMsg(conn.Vector),
			Ds:     int64(conn.ConnDS),
		}
	}
	Log.Debug("sending all connections", len(msg.AllConnections))

	msg.AllTrios = make([]*m3api.TrioMsg, len(ppd.AllTrioDetails))
	for idx, tr := range ppd.AllTrioDetails {
		msg.AllTrios[idx] = &m3api.TrioMsg{TrioId: int32(tr.GetId()), ConnIds: []int32{
			int32(tr.GetConnections()[0].GetId()),
			int32(tr.GetConnections()[1].GetId()),
			int32(tr.GetConnections()[2].GetId()),
		}}
	}
	Log.Debug("sending all trios", len(msg.AllTrios))

	msg.ValidNextTrioIds = make([]int32, 12*2)
	for idx, pair := range ppd.GetValidNextTrio() {
		msg.ValidNextTrioIds[idx*2] = int32(pair[0])
		msg.ValidNextTrioIds[idx*2+1] = int32(pair[1])
	}
	msg.Mod4PermutationsTrioIds = make([]int32, 12*4)
	for idx, quad := range ppd.GetAllMod4Permutations() {
		for k := 0; k < 4; k++ {
			msg.Mod4PermutationsTrioIds[idx*4+k] = int32(quad[k])
		}
	}
	msg.Mod8PermutationsTrioIds = make([]int32, 12*8)
	for idx, eight := range ppd.GetAllMod8Permutations() {
		for k := 0; k < 8; k++ {
			msg.Mod8PermutationsTrioIds[idx*8+k] = int32(eight[k])
		}
	}
	Log.Debug("sending all valid trios and permutations")

	msg.AllGrowthContexts = make([]*m3api.GrowthContextMsg, len(ppd.AllGrowthContexts))
	for idx, gc := range ppd.AllGrowthContexts {
		msg.AllGrowthContexts[idx] = &m3api.GrowthContextMsg{
			GrowthContextId: int32(gc.GetId()),
			GrowthType:      int32(gc.GetGrowthType()),
			GrowthIndex:     int32(gc.GetGrowthIndex()),
		}
	}
	Log.Debug("sending all growth context", len(msg.AllGrowthContexts))

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
