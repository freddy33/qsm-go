package pointdb

import (
	"github.com/freddy33/qsm-go/model/m3point"
)

func (pointData *ServerPointPackData) InitializeAll() {
	pointData.initConnections()
	pointData.initTrioDetails()
	pointData.initGrowthContexts()
	pointData.initContextCubes()
	pointData.initPathBuilders()
}

func (pointData *ServerPointPackData) GetValidNextTrio() [12][2]m3point.TrioIndex {
	return validNextTrio
}

func (pointData *ServerPointPackData) GetAllMod4Permutations() [12][4]m3point.TrioIndex {
	return allMod4Permutations
}

func (pointData *ServerPointPackData) GetAllMod8Permutations() [12][8]m3point.TrioIndex {
	return allMod8Permutations
}

func (pointData *ServerPointPackData) initConnections() {
	if !pointData.ConnectionsLoaded {
		err := pointData.loadConnectionDetails()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d connection details", pointData.GetEnvId(), len(pointData.AllConnections))
	}
}

func (pointData *ServerPointPackData) initTrioDetails() {
	if !pointData.TrioDetailsLoaded {
		err := pointData.loadTrioDetails()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d trio details", pointData.GetEnvId(), len(pointData.AllTrioDetails))
	}
}

func (pointData *ServerPointPackData) initGrowthContexts() {
	if !pointData.GrowthContextsLoaded {
		err := pointData.loadGrowthContexts()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d growth contexts", pointData.GetEnvId(), len(pointData.AllGrowthContexts))
	}
}

func (pointData *ServerPointPackData) initContextCubes() {
	if !pointData.cubesLoaded {
		err := pointData.loadContextCubes()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d cubes", pointData.GetEnvId(), len(pointData.cubeIdsPerKey))
	}
}

func (pointData *ServerPointPackData) initPathBuilders() {
	if !pointData.pathBuildersLoaded {
		err := pointData.loadPathBuilders()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d path builders", pointData.GetEnvId(), len(pointData.pathBuilders))
	}
}

func (pointData *ServerPointPackData) FillDb() {
	n, err := pointData.saveAllConnectionDetails()
	if err != nil {
		Log.Fatalf("could not save all connections due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d connection details", pointData.GetEnvId(), n)
	}

	pointData.initConnections()

	n, err = pointData.saveAllTrioDetails()
	if err != nil {
		Log.Fatalf("could not save all trios due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio details", pointData.GetEnvId(), n)
	}
	pointData.initTrioDetails()

	n, err = pointData.saveAllGrowthContexts()
	if err != nil {
		Log.Fatalf("could not save all growth contexts due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d growth contexts", pointData.GetEnvId(), n)
	}
	pointData.initGrowthContexts()

	n, err = pointData.saveAllContextCubes()
	if err != nil {
		Log.Fatalf("could not save all contexts cubes due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d contexts cubes", pointData.GetEnvId(), n)
	}
	pointData.initContextCubes()

	n, err = pointData.saveAllPathBuilders()
	if err != nil {
		Log.Fatalf("could not save all path builders due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d path builders", pointData.GetEnvId(), n)
	}
	pointData.initPathBuilders()
}
