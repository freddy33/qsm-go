package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/backend/pathdb"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
)

func createPathContext(w http.ResponseWriter, r *http.Request) {
	Log.Infof("Receive createPathContext")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Log.Errorf("receive wrong message in createPathContext: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be read")
		return
	}
	pMsg := &m3api.PathContextMsg{}
	err = proto.Unmarshal(b, pMsg)
	if err != nil {
		Log.Errorf("Could not parse body in createPathContext: %s", err.Error())
		SendResponse(w, http.StatusBadRequest, "req body could not be parsed")
		return
	}
	env := GetEnvironment(r)
	pointData := pointdb.GetPointPackData(env)
	ppd := pathdb.GetServerPathPackData(env)
	newPathCtx := ppd.CreatePathCtxFromAttributes(
		pointData.GetGrowthContextById(int(pMsg.GetGrowthContextId())),
		int(pMsg.GetGrowthOffset()),
		m3api.PointMsgToPoint(pMsg.Center))

	resMsg := m3api.PathContextMsg{
		PathCtxId:       int32(newPathCtx.GetId()),
		GrowthContextId: int32(newPathCtx.GetId()),
		GrowthOffset:    int32(newPathCtx.GetGrowthOffset()),
		// TODO: Uncomment the one line below once init node done
		//Center: m3api.PointToPointMsg(newPathCtx.GetRootPathNode().P()),
		Center: pMsg.Center,
	}

	data, err := proto.Marshal(&resMsg)
	if err != nil {
		Log.Warnf("Failed to marshal PathContextMsg due to: %q", err.Error())
		w.WriteHeader(500)
		_, err = fmt.Fprintf(w, "Failed to marshal PathContextMsg due to:\n%s\n", err.Error())
		if err != nil {
			Log.Errorf("failed to send error message to response due to %q", err.Error())
		}
	}
	_, err = w.Write(data)
	if err != nil {
		Log.Errorf("failed to send data to response due to %q", err.Error())
	}
}
