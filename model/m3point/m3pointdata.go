package m3point

import (
	"github.com/freddy33/qsm-go/m3util"
)

type PointPackDataIfc interface {
	m3util.QsmDataPack
	GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails
	GetConnDetailsByVector(vector Point) *ConnectionDetails
	GetGrowthContextByTypeAndIndex(growthType GrowthType, index int) GrowthContext
	GetAllTrioDetails() []*TrioDetails
	GetTrioDetails(trIdx TrioIndex) *TrioDetails
	GetAllConnDetailsByVector() map[Point]*ConnectionDetails
	GetConnDetailsById(id ConnectionId) *ConnectionDetails
	GetPathNodeBuilder(growthCtx GrowthContext, offset int, c Point) PathNodeBuilder
	GetValidNextTrio() [12][2]TrioIndex
	GetAllMod4Permutations() [12][4]TrioIndex
	GetAllMod8Permutations() [12][8]TrioIndex

	GetNbPathBuilders() int
	GetMaxConnId() ConnectionId
	GetAllGrowthContexts() []GrowthContext
	GetGrowthContextById(id int) GrowthContext
	GetPathNodeBuilderById(cubeId int) PathNodeBuilder
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

func (ppd *BasePointPackData) CheckPathBuildersInitialized() {
	if !ppd.PathBuildersLoaded {
		Log.Fatalf("Path Builders should have been initialized! Please call m3point.InitializeDBEnv(envId=%d) method before this!", ppd.GetEnvId())
	}
}

func (ppd *BasePointPackData) GetNbPathBuilders() int {
	ppd.CheckPathBuildersInitialized()
	return len(ppd.PathBuilders)
}

