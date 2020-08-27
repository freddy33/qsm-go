package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/golang/protobuf/proto"
	"net/http"
)

var Log = m3util.NewLogger("m3server", m3util.INFO)

func retrievePointData(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive retrievePointData")

	w.Header().Set("Content-Type", "application/x-protobuf; messageType=backend.m3api.PointPackDataMsg")

	env := GetEnvironment(r)
	ppd, _ := getServerPointPackData(env)
	msg := m3api.PointPackDataMsg{}

	msg.AllConnections = make([]*m3api.ConnectionMsg, len(ppd.AllConnections))
	for idx, conn := range ppd.AllConnections {
		msg.AllConnections[idx] = &m3api.ConnectionMsg{
			ConnId: int32(conn.GetId()),
			Vector: &m3api.PointMsg{X: int32(conn.Vector.X()), Y: int32(conn.Vector.Y()), Z: int32(conn.Vector.Z())},
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
	for idx, pair := range validNextTrio {
		msg.ValidNextTrioIds[idx*2] = int32(pair[0])
		msg.ValidNextTrioIds[idx*2+1] = int32(pair[1])
	}
	msg.Mod4PermutationsTrioIds = make([]int32, 12*4)
	for idx, quad := range AllMod4Permutations {
		for k := 0; k < 4; k++ {
			msg.Mod4PermutationsTrioIds[idx*4+k] = int32(quad[k])
		}
	}
	msg.Mod8PermutationsTrioIds = make([]int32, 12*8)
	for idx, eight := range AllMod8Permutations {
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

	msg.AllCubes = make([]*m3api.CubeOfTrioMsg, len(ppd.CubeIdsPerKey)+1)
	// Dummy 0 cube
	msg.AllCubes[0] = &m3api.CubeOfTrioMsg{CubeId: int32(0)}
	for cubeKey, id := range ppd.CubeIdsPerKey {
		cubeOfTrio := cubeKey.GetCube()
		centerFaces := cubeOfTrio.GetCenterFaces()
		middleEdges := cubeOfTrio.GetMiddleEdges()
		msg.AllCubes[id] = &m3api.CubeOfTrioMsg{
			CubeId:             int32(id),
			GrowthContextId:    int32(cubeKey.GetGrowthCtxId()),
			CenterTrioId:       int32(cubeOfTrio.GetCenter()),
			CenterFacesTrioIds: convertToInt32Slice(centerFaces[:]),
			MiddleEdgesTrioIds: convertToInt32Slice(middleEdges[:]),
		}
	}
	Log.Debug("sending all cubes", len(msg.AllCubes))

	msg.AllPathNodeBuilders = make([]*m3api.RootPathNodeBuilderMsg, len(ppd.PathBuilders))
	for idx, pb := range ppd.PathBuilders {
		if idx == 0 {
			// Dummy 0 cube Id
			msg.AllPathNodeBuilders[idx] = &m3api.RootPathNodeBuilderMsg{
				CubeId: int32(0),
				TrioId: int32(m3point.NilTrioIndex),
			}
		} else {
			msg.AllPathNodeBuilders[idx] = &m3api.RootPathNodeBuilderMsg{
				CubeId:            int32(pb.GetCubeId()),
				TrioId:            int32(pb.GetTrioIndex()),
				InterNodeBuilders: convertToInterMsg(pb.GetPathLinks()),
			}
		}
	}
	Log.Debug("sending all root path builders", len(msg.AllPathNodeBuilders))

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

func convertToInt32Slice(trIds []m3point.TrioIndex) []int32 {
	res := make([]int32, len(trIds))
	for idx, tr := range trIds {
		res[idx] = int32(tr)
	}
	return res
}

func convertToLastMsg(pnb m3point.PathNodeBuilder) *m3api.LastPathNodeBuilderMsg {
	lpnb := pnb.(*m3point.LastPathNodeBuilder)
	return &m3api.LastPathNodeBuilderMsg{
		CubeId:          int32(lpnb.GetCubeId()),
		TrioId:          int32(lpnb.GetTrioIndex()),
		NextMainConnId:  int32(lpnb.GetNextMainConnId()),
		NextInterConnId: int32(lpnb.GetNextInterConnId()),
	}
}

func convertToInterMsg(pls []m3point.PathLinkBuilder) []*m3api.IntermediatePathNodeBuilderMsg {
	res := make([]*m3api.IntermediatePathNodeBuilderMsg, len(pls))
	for idx, pl := range pls {
		pnb := pl.GetPathNodeBuilder().(*m3point.IntermediatePathNodeBuilder)
		nextPls := pnb.GetPathLinks()
		res[idx] = &m3api.IntermediatePathNodeBuilderMsg{
			CubeId:           int32(pnb.GetCubeId()),
			TrioId:           int32(pnb.GetTrioIndex()),
			Link1ConnId:      int32(nextPls[0].GetConnectionId()),
			LastNodeBuilder1: convertToLastMsg(nextPls[0].GetPathNodeBuilder()),
			Link2ConnId:      int32(nextPls[1].GetConnectionId()),
			LastNodeBuilder2: convertToLastMsg(nextPls[1].GetPathNodeBuilder()),
		}
	}
	return res
}
