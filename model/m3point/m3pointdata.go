package m3point

import "github.com/freddy33/qsm-go/utils/m3db"

type PointData interface {
	GetEnvId() m3db.QsmEnvID
	GetMaxConnId() ConnectionId
	GetConnDetailsById(id ConnectionId) *ConnectionDetails
	GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails
	GetGrowthContextById(id int) GrowthContext
	GetGrowthContextByTypeAndIndex(growthType GrowthType, index int) GrowthContext
	GetPathNodeBuilder(growthCtx GrowthContext, offset int, c Point) PathNodeBuilder
	GetPathNodeBuilderById(cubeId int) PathNodeBuilder
	GetTrioDetails(trIdx TrioIndex) *TrioDetails
	GetTrioTransitionTableTxt() map[Int2][7]string
	GetTrioTableCsv() [][]string
	GetCubeById(cubeId int) CubeKeyId
	GetCubeIdByKey(cubeKey CubeKeyId) int
}

type PointPackData struct {
	Env *m3db.QsmEnvironment

	// All connection details ordered and mapped by base vector
	AllConnections         []*ConnectionDetails
	AllConnectionsByVector map[Point]*ConnectionDetails
	ConnectionsLoaded      bool

	// All the possible trio details used
	AllTrioDetails    TrioDetailList
	TrioDetailsLoaded bool

	// Collection of all growth context ordered
	AllGrowthContexts    []GrowthContext
	GrowthContextsLoaded bool

	// Identified all cubes uniquely
	CubeIdsPerKey map[CubeKeyId]int
	CubesLoaded   bool

	// The index of this slice is the cube id
	PathBuilders       []*RootPathNodeBuilder
	PathBuildersLoaded bool
}

func GetPointPackData(env *m3db.QsmEnvironment) *PointPackData {
	if env.GetData(m3db.PointIdx) == nil {
		ppd := new(PointPackData)
		ppd.Env = env
		env.SetData(m3db.PointIdx, ppd)
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3db.PointIdx).(*PointPackData)
}

func (ppd *PointPackData) GetEnvId() m3db.QsmEnvID {
	if ppd == nil {
		return m3db.NoEnv
	}
	if ppd.Env == nil {
		return m3db.NoEnv
	}
	return ppd.Env.GetId()
}

func (ppd *PointPackData) resetFlags() {
	ppd.ConnectionsLoaded = false
	ppd.TrioDetailsLoaded = false
	ppd.GrowthContextsLoaded = false
	ppd.CubesLoaded = false
	ppd.PathBuildersLoaded = false
}

func (ppd *PointPackData) checkConnInitialized() {
	if !ppd.ConnectionsLoaded {
		Log.Fatalf("Connections should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *PointPackData) checkTrioInitialized() {
	if !ppd.TrioDetailsLoaded {
		Log.Fatalf("Trios should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *PointPackData) checkGrowthContextsInitialized() {
	if !ppd.TrioDetailsLoaded {
		Log.Fatalf("trio contexts should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *PointPackData) checkCubesInitialized() {
	if !ppd.CubesLoaded {
		Log.Fatalf("Cubes should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *PointPackData) checkPathBuildersInitialized() {
	if !ppd.PathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}
