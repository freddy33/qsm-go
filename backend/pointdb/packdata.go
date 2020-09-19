package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ServerPointPackDataIfc interface {
	m3point.PointPackDataIfc

	GetConnDetailsByVector(vector m3point.Point) *m3point.ConnectionDetails

	GetAllTrioDetails() []*m3point.TrioDetails
	GetTrioDetails(trIdx m3point.TrioIndex) *m3point.TrioDetails
	GetAllConnDetailsByVector() map[m3point.Point]*m3point.ConnectionDetails
	GetPathNodeBuilder(growthCtx m3point.GrowthContext, offset int, c m3point.Point) PathNodeBuilder

	GetNbPathBuilders() int
	GetPathNodeBuilderById(cubeId int) PathNodeBuilder
	GetCubeById(cubeId int) CubeKeyId
	GetCubeIdByKey(cubeKey CubeKeyId) int
}

type ServerPointPackData struct {
	m3point.BasePointPackData
	Env *m3db.QsmDbEnvironment

	// Identified all cubes uniquely
	CubeIdsPerKey map[CubeKeyId]int
	CubesLoaded   bool

	// The index of this slice is the cube id
	PathBuilders       []*RootPathNodeBuilder
	PathBuildersLoaded bool
}

func GetPointPackData(env m3util.QsmEnvironment) ServerPointPackDataIfc {
	ppd, _ := GetServerPointPackData(env)
	return ppd
}

func GetServerPointPackData(env m3util.QsmEnvironment) (*ServerPointPackData, bool) {
	newData := env.GetData(m3util.PointIdx) == nil
	if newData {
		ppd := new(ServerPointPackData)
		ppd.EnvId = env.GetId()
		ppd.Env = env.(*m3db.QsmDbEnvironment)
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in env data array
	}
	return env.GetData(m3util.PointIdx).(*ServerPointPackData), newData
}

func (ppd *ServerPointPackData) ResetFlags() {
	ppd.ConnectionsLoaded = false
	ppd.TrioDetailsLoaded = false
	ppd.GrowthContextsLoaded = false
	ppd.CubesLoaded = false
	ppd.PathBuildersLoaded = false
}

func (ppd *ServerPointPackData) CheckCubesInitialized() {
	if !ppd.CubesLoaded {
		Log.Fatalf("Cubes should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *ServerPointPackData) CheckPathBuildersInitialized() {
	if !ppd.PathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *ServerPointPackData) GetNbPathBuilders() int {
	ppd.CheckPathBuildersInitialized()
	return len(ppd.PathBuilders)
}

func (ppd *ServerPointPackData) GetPathNodeBuilderById(cubeId int) PathNodeBuilder {
	return ppd.PathBuilders[cubeId]
}

func (ppd *ServerPointPackData) GetCubeById(cubeId int) CubeKeyId {
	ppd.CheckCubesInitialized()
	for cubeKey, id := range ppd.CubeIdsPerKey {
		if id == cubeId {
			return cubeKey
		}
	}
	Log.Fatalf("trying to find cube by id %d which does not exists", cubeId)
	return CubeKeyId{-1, CubeOfTrioIndex{}}
}

func (ppd *ServerPointPackData) GetCubeIdByKey(cubeKey CubeKeyId) int {
	ppd.CheckCubesInitialized()
	id, ok := ppd.CubeIdsPerKey[cubeKey]
	if !ok {
		Log.Fatalf("trying to find cube %v which does not exists", cubeKey)
		return -1
	}
	return id
}

func (ppd *ServerPointPackData) GetPathNodeBuilder(growthCtx m3point.GrowthContext, offset int, c m3point.Point) PathNodeBuilder {
	ppd.CheckPathBuildersInitialized()
	// TODO: Verify the key below stay local and is not staying in memory
	key := CubeKeyId{GrowthCtxId: growthCtx.GetId(), Cube: CreateTrioCube(ppd, growthCtx, offset, c)}
	cubeId := ppd.GetCubeIdByKey(key)
	return ppd.GetPathNodeBuilderById(cubeId)
}
