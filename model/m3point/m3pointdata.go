package m3point

import "github.com/freddy33/qsm-go/utils/m3db"

type PointPackData struct {
	env *m3db.QsmEnvironment

	// All connection details ordered and mapped by base vector
	allConnections         []*ConnectionDetails
	allConnectionsByVector map[Point]*ConnectionDetails
	connectionsLoaded      bool

	// All the possible trio details used
	allTrioDetails    TrioDetailList
	trioDetailsLoaded bool

	// Collection of all growth context ordered
	allGrowthContexts    []GrowthContext
	growthContextsLoaded bool

	// Identified all cubes uniquely
	cubeIdsPerKey map[CubeKeyId]int
	cubesLoaded   bool

	// The index of this slice is the cube id
	pathBuilders       []*RootPathNodeBuilder
	pathBuildersLoaded bool
}

func GetPointPackData(env *m3db.QsmEnvironment) *PointPackData {
	if env.GetData(m3db.PointIdx) == nil {
		ppd := new(PointPackData)
		ppd.env = env
		env.SetData(m3db.PointIdx, ppd)
		// do not return ppd but always the pointer in env data array
	}
	return env.GetData(m3db.PointIdx).(*PointPackData)
}

func (ppd *PointPackData) GetId() m3db.QsmEnvID {
	if ppd == nil {
		return m3db.NoEnv
	}
	if ppd.env == nil {
		return m3db.NoEnv
	}
	return ppd.env.GetId()
}

func (ppd *PointPackData) resetFlags() {
	ppd.connectionsLoaded = false
	ppd.trioDetailsLoaded = false
	ppd.growthContextsLoaded = false
	ppd.cubesLoaded = false
	ppd.pathBuildersLoaded = false
}

func (ppd *PointPackData) checkConnInitialized() {
	if !ppd.connectionsLoaded {
		Log.Fatalf("Connections should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetId())
	}
}

func (ppd *PointPackData) checkTrioInitialized() {
	if !ppd.trioDetailsLoaded {
		Log.Fatalf("Trios should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetId())
	}
}

func (ppd *PointPackData) checkGrowthContextsInitialized() {
	if !ppd.trioDetailsLoaded {
		Log.Fatalf("trio contexts should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetId())
	}
}

func (ppd *PointPackData) checkCubesInitialized() {
	if !ppd.cubesLoaded {
		Log.Fatalf("Cubes should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetId())
	}
}

func (ppd *PointPackData) checkPathBuildersInitialized() {
	if !ppd.pathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetId())
	}
}
