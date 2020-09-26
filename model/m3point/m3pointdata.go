package m3point

import (
	"github.com/freddy33/qsm-go/m3util"
)

type PointPackDataIfc interface {
	m3util.QsmDataPack

	// Used by m3gl
	GetMaxConnId() ConnectionId
	GetConnDetailsById(id ConnectionId) *ConnectionDetails

	// Used by m3space
	// Used in m3space.m3node
	GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails
	// Used by m3space.m3event
	GetGrowthContextByTypeAndIndex(growthType GrowthType, index int) GrowthContext
	GetGrowthContextById(id int) GrowthContext

	// Follow should be used by UI
	GetAllGrowthContexts() []GrowthContext

	// Used by space tests
	GetAllMod8Permutations() [12][8]TrioIndex
	GetValidNextTrio() [12][2]TrioIndex
	GetAllMod4Permutations() [12][4]TrioIndex
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
}

func (ppd *BasePointPackData) GetEnvId() m3util.QsmEnvID {
	if ppd == nil {
		return m3util.NoEnv
	}
	return ppd.EnvId
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

func (ppd *BasePointPackData) GetMaxConnId() ConnectionId {
	ppd.CheckConnInitialized()
	// The pos conn Id of the last one
	return ppd.AllConnections[len(ppd.AllConnections)-1].GetPosId()
}

func (ppd *BasePointPackData) GetConnDetailsById(id ConnectionId) *ConnectionDetails {
	ppd.CheckConnInitialized()
	if id > 0 {
		return ppd.AllConnections[2*id-2]
	} else {
		return ppd.AllConnections[-2*id-1]
	}
}

func (ppd *BasePointPackData) GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails {
	return ppd.GetConnDetailsByVector(MakeVector(p1, p2))
}

func (ppd *BasePointPackData) GetConnDetailsByVector(vector Point) *ConnectionDetails {
	ppd.CheckConnInitialized()
	cd, ok := ppd.AllConnectionsByVector[vector]
	if !ok {
		Log.Error("Vector", vector, "is not a known connection details")
		return &EmptyConnDetails
	}
	return cd
}

func (ppd *BasePointPackData) GetAllConnDetailsByVector() map[Point]*ConnectionDetails {
	ppd.CheckConnInitialized()
	return ppd.AllConnectionsByVector
}

func (ppd *BasePointPackData) GetAllGrowthContexts() []GrowthContext {
	ppd.checkGrowthContextsInitialized()
	return ppd.AllGrowthContexts
}

func (ppd *BasePointPackData) GetGrowthContextById(id int) GrowthContext {
	ppd.checkGrowthContextsInitialized()
	return ppd.AllGrowthContexts[id]
}

func (ppd *BasePointPackData) GetGrowthContextByTypeAndIndex(growthType GrowthType, index int) GrowthContext {
	ppd.checkGrowthContextsInitialized()
	for _, growthCtx := range ppd.AllGrowthContexts {
		if growthCtx.GetGrowthType() == growthType && growthCtx.GetGrowthIndex() == index {
			return growthCtx
		}
	}
	Log.Fatalf("could not find trio Context for %d %d", growthType, index)
	return nil
}

func (ppd *BasePointPackData) GetAllTrioDetails() []*TrioDetails {
	ppd.checkTrioInitialized()
	return ppd.AllTrioDetails
}

func (ppd *BasePointPackData) GetTrioDetails(trIdx TrioIndex) *TrioDetails {
	ppd.checkTrioInitialized()
	return ppd.AllTrioDetails[trIdx]
}

