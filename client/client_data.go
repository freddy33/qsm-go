package client

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ClientPointPackData struct {
	m3point.BasePointPackData
	ValidNextTrio       [12][2]m3point.TrioIndex
	AllMod4Permutations [12][4]m3point.TrioIndex
	AllMod8Permutations [12][8]m3point.TrioIndex
}

type ClientPathPackData struct {
	m3path.BasePathPackData
	clConn *ClientConnection
}

func (cl *ClientConnection) GetClientPointPackData(env m3util.QsmEnvironment) *ClientPointPackData {
	if env.GetData(m3util.PointIdx) == nil {
		ppd := new(ClientPointPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3util.PointIdx).(*ClientPointPackData)
}

func (cl *ClientConnection) GetClientPathPackData(env m3util.QsmEnvironment) *ClientPathPackData {
	if env.GetData(m3util.PathIdx) == nil {
		ppd := new(ClientPathPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PathIdx, ppd)
		ppd.clConn = cl
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3util.PathIdx).(*ClientPathPackData)
}

func (ppd *ClientPointPackData) GetValidNextTrio() [12][2]m3point.TrioIndex {
	return ppd.ValidNextTrio
}

func (ppd *ClientPointPackData) GetAllMod4Permutations() [12][4]m3point.TrioIndex {
	return ppd.AllMod4Permutations
}

func (ppd *ClientPointPackData) GetAllMod8Permutations() [12][8]m3point.TrioIndex {
	return ppd.AllMod8Permutations
}

func (ppd *ClientPointPackData) GetPathNodeBuilder(growthCtx m3point.GrowthContext, offset int, c m3point.Point) m3point.PathNodeBuilder {
	ppd.CheckPathBuildersInitialized()
	// TODO: Verify the key below stay local and is not staying in memory
	key := m3point.CubeKeyId{GrowthCtxId: growthCtx.GetId(), Cube: m3point.CreateTrioCube(ppd, growthCtx, offset, c)}
	cubeId := ppd.GetCubeIdByKey(key)
	return ppd.GetPathNodeBuilderById(cubeId)
}
