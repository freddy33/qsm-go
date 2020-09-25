package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ServerPointPackData struct {
	m3point.BasePointPackData
	env *m3db.QsmDbEnvironment

	// Identified all cubes uniquely
	cubeIdsPerKey map[CubeKeyId]int
	cubesLoaded   bool

	// The index of this slice is the cube id
	pathBuilders       []*RootPathNodeBuilder
	pathBuildersLoaded bool

	connDetailsTe  *m3db.TableExec
	trioDetailsTe  *m3db.TableExec
	growthCtxTe    *m3db.TableExec
	trioCubesTe    *m3db.TableExec
	pathBuildersTe *m3db.TableExec
}

func GetPointPackData(env m3util.QsmEnvironment) *ServerPointPackData {
	if env.GetData(m3util.PointIdx) == nil {
		ppd := new(ServerPointPackData)
		ppd.EnvId = env.GetId()
		ppd.env = env.(*m3db.QsmDbEnvironment)
		env.SetData(m3util.PointIdx, ppd)
	}
	// do not return ppd but always the pointer in env data array
	return env.GetData(m3util.PointIdx).(*ServerPointPackData)
}

func (ppd *ServerPointPackData) ResetFlags() {
	ppd.ConnectionsLoaded = false
	ppd.TrioDetailsLoaded = false
	ppd.GrowthContextsLoaded = false
	ppd.cubesLoaded = false
	ppd.pathBuildersLoaded = false
}

func (ppd *ServerPointPackData) CheckCubesInitialized() {
	if !ppd.cubesLoaded {
		Log.Fatalf("Cubes should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *ServerPointPackData) CheckPathBuildersInitialized() {
	if !ppd.pathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *ServerPointPackData) GetNbPathBuilders() int {
	ppd.CheckPathBuildersInitialized()
	return len(ppd.pathBuilders)
}

func (ppd *ServerPointPackData) GetPathNodeBuilderById(cubeId int) PathNodeBuilder {
	return ppd.pathBuilders[cubeId]
}

func (ppd *ServerPointPackData) GetCubeById(cubeId int) CubeKeyId {
	ppd.CheckCubesInitialized()
	for cubeKey, id := range ppd.cubeIdsPerKey {
		if id == cubeId {
			return cubeKey
		}
	}
	Log.Fatalf("trying to find cube by id %d which does not exists", cubeId)
	return CubeKeyId{-1, CubeOfTrioIndex{}}
}

func (ppd *ServerPointPackData) GetCubeIdByKey(cubeKey CubeKeyId) int {
	ppd.CheckCubesInitialized()
	id, ok := ppd.cubeIdsPerKey[cubeKey]
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

func (ppd *ServerPointPackData) createTables() {
	var err error
	ppd.connDetailsTe, err = ppd.env.GetOrCreateTableExec(ConnectionDetailsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", ConnectionDetailsTable, err)
		return
	}
	ppd.trioDetailsTe, err = ppd.env.GetOrCreateTableExec(TrioDetailsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", TrioDetailsTable, err)
		return
	}
	ppd.growthCtxTe, err = ppd.env.GetOrCreateTableExec(GrowthContextsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", GrowthContextsTable, err)
		return
	}
	ppd.trioCubesTe, err = ppd.env.GetOrCreateTableExec(TrioCubesTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", TrioCubesTable, err)
		return
	}
	ppd.pathBuildersTe, err = ppd.env.GetOrCreateTableExec(PathBuildersTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PathBuildersTable, err)
		return
	}
}
