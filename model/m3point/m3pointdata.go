package m3point

import (
	"github.com/freddy33/qsm-go/utils/m3util"
)

type PointData interface {
	GetEnvId() m3util.QsmEnvID
	GetMaxConnId() ConnectionId
	GetConnDetailsById(id ConnectionId) *ConnectionDetails
	GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails
	GetGrowthContextById(id int) GrowthContext
	GetGrowthContextByTypeAndIndex(growthType GrowthType, index int) GrowthContext
	GetPathNodeBuilder(growthCtx GrowthContext, offset int, c Point) PathNodeBuilder
	GetPathNodeBuilderById(cubeId int) PathNodeBuilder
	GetTrioDetails(trIdx TrioIndex) *TrioDetails
	GetTrioTableCsv() [][]string
	GetCubeById(cubeId int) CubeKeyId
	GetCubeIdByKey(cubeKey CubeKeyId) int
}

type BasePointPackData struct {
	EnvId m3util.QsmEnvID

	// All connection details ordered and mapped by base vector
	AllConnections         []*ConnectionDetails
	AllConnectionsByVector map[Point]*ConnectionDetails
	ConnectionsLoaded      bool

	// All the possible trio details used
	AllTrioDetails      []*TrioDetails
	ValidNextTrio       [12][2]TrioIndex
	AllMod4Permutations [12][4]TrioIndex
	AllMod8Permutations [12][8]TrioIndex
	TrioDetailsLoaded   bool

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

func (ppd *BasePointPackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	return ppd.EnvId
}

func GetPointPackData(env m3util.QsmEnvironment) *BasePointPackData {
	if env.GetData(m3util.PointIdx) == nil {
		ppd := new(BasePointPackData)
		ppd.EnvId = env.GetId()
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in Env data array
	}
	return env.GetData(m3util.PointIdx).(*BasePointPackData)
}

func (ppd *BasePointPackData) ResetFlags() {
	ppd.ConnectionsLoaded = false
	ppd.TrioDetailsLoaded = false
	ppd.GrowthContextsLoaded = false
	ppd.CubesLoaded = false
	ppd.PathBuildersLoaded = false
}

func (ppd *BasePointPackData) CheckConnInitialized() {
	if !ppd.ConnectionsLoaded {
		Log.Fatalf("Connections should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *BasePointPackData) checkTrioInitialized() {
	if !ppd.TrioDetailsLoaded {
		Log.Fatalf("Trios should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *BasePointPackData) checkGrowthContextsInitialized() {
	if !ppd.TrioDetailsLoaded {
		Log.Fatalf("trio contexts should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *BasePointPackData) CheckCubesInitialized() {
	if !ppd.CubesLoaded {
		Log.Fatalf("Cubes should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *BasePointPackData) checkPathBuildersInitialized() {
	if !ppd.PathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}
