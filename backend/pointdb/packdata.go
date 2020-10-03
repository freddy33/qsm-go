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

func GetServerPointPackData(env m3util.QsmEnvironment) *ServerPointPackData {
	if env.GetData(m3util.PointIdx) == nil {
		ppd := new(ServerPointPackData)
		ppd.EnvId = env.GetId()
		ppd.env = env.(*m3db.QsmDbEnvironment)
		env.SetData(m3util.PointIdx, ppd)
	}
	// do not return ppd but always the pointer in env data array
	return env.GetData(m3util.PointIdx).(*ServerPointPackData)
}

func (pointData *ServerPointPackData) ResetFlags() {
	pointData.ConnectionsLoaded = false
	pointData.TrioDetailsLoaded = false
	pointData.GrowthContextsLoaded = false
	pointData.cubesLoaded = false
	pointData.pathBuildersLoaded = false
}

func (pointData *ServerPointPackData) CheckCubesInitialized() {
	if !pointData.cubesLoaded {
		Log.Fatalf("Cubes should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", pointData.GetEnvId())
	}
}

func (pointData *ServerPointPackData) CheckPathBuildersInitialized() {
	if !pointData.pathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", pointData.GetEnvId())
	}
}

func (pointData *ServerPointPackData) GetNbPathBuilders() int {
	pointData.CheckPathBuildersInitialized()
	return len(pointData.pathBuilders)
}

func (pointData *ServerPointPackData) GetRootPathNodeBuilderById(cubeId int) PathNodeBuilder {
	return pointData.pathBuilders[cubeId]
}

func (pointData *ServerPointPackData) GetCubeById(cubeId int) CubeKeyId {
	pointData.CheckCubesInitialized()
	for cubeKey, id := range pointData.cubeIdsPerKey {
		if id == cubeId {
			return cubeKey
		}
	}
	Log.Fatalf("trying to find cube by id %d which does not exists", cubeId)
	return CubeKeyId{-1, CubeOfTrioIndex{}}
}

func (pointData *ServerPointPackData) GetCubeIdByKey(cubeKey CubeKeyId) int {
	pointData.CheckCubesInitialized()
	id, ok := pointData.cubeIdsPerKey[cubeKey]
	if !ok {
		Log.Fatalf("trying to find cube %v which does not exists", cubeKey)
		return -1
	}
	return id
}

func (pointData *ServerPointPackData) GetPathNodeBuilder(growthCtx m3point.GrowthContext, offset int, c m3point.Point) PathNodeBuilder {
	pointData.CheckPathBuildersInitialized()
	// TODO: Verify the key below stay local and is not staying in memory
	key := CubeKeyId{GrowthCtxId: growthCtx.GetId(), Cube: CreateTrioCube(pointData, growthCtx, offset, c)}
	cubeId := pointData.GetCubeIdByKey(key)
	return pointData.GetRootPathNodeBuilderById(cubeId)
}

func (pointData *ServerPointPackData) createTables() {
	tableNames := [5]string{ConnectionDetailsTable, TrioDetailsTable, GrowthContextsTable, TrioCubesTable, PathBuildersTable}
	pointTableExecs := [5]*m3db.TableExec{}

	// IMPORTANT: Create ALL the tables before preparing the queries
	var err error

	for i := 0; i < len(pointTableExecs); i++ {
		pointTableExecs[i], err = pointData.env.GetOrCreateTableExec(tableNames[i])
		if err != nil {
			Log.Fatal(err)
			return
		}
	}

	for i := 0; i < len(pointTableExecs); i++ {
		err = pointTableExecs[i].PrepareQueries()
		if err != nil {
			Log.Fatal(err)
			return
		}
	}

	pointData.connDetailsTe = pointTableExecs[0]
	pointData.trioDetailsTe = pointTableExecs[1]
	pointData.growthCtxTe = pointTableExecs[2]
	pointData.trioCubesTe = pointTableExecs[3]
	pointData.pathBuildersTe = pointTableExecs[4]
}
