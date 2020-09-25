package pointdb

import (
	"github.com/freddy33/qsm-go/model/m3point"
)

func (ppd *ServerPointPackData) InitializeAll() {
	ppd.initConnections()
	ppd.initTrioDetails()
	ppd.initGrowthContexts()
	ppd.initContextCubes()
	ppd.initPathBuilders()
}

func (ppd *ServerPointPackData) GetValidNextTrio() [12][2]m3point.TrioIndex {
	return validNextTrio
}

func (ppd *ServerPointPackData) GetAllMod4Permutations() [12][4]m3point.TrioIndex {
	return allMod4Permutations
}

func (ppd *ServerPointPackData) GetAllMod8Permutations() [12][8]m3point.TrioIndex {
	return allMod8Permutations
}

func (ppd *ServerPointPackData) initConnections() {
	if !ppd.ConnectionsLoaded {
		err := ppd.loadConnectionDetails()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d connection details", ppd.GetEnvId(), len(ppd.AllConnections))
	}
}

func (ppd *ServerPointPackData) initTrioDetails() {
	if !ppd.TrioDetailsLoaded {
		err := ppd.loadTrioDetails()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d trio details", ppd.GetEnvId(), len(ppd.AllTrioDetails))
	}
}

func (ppd *ServerPointPackData) initGrowthContexts() {
	if !ppd.GrowthContextsLoaded {
		err := ppd.loadGrowthContexts()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d growth contexts", ppd.GetEnvId(), len(ppd.AllGrowthContexts))
	}
}

func (ppd *ServerPointPackData) initContextCubes() {
	if !ppd.cubesLoaded {
		err := ppd.loadContextCubes()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d cubes", ppd.GetEnvId(), len(ppd.cubeIdsPerKey))
	}
}

func (ppd *ServerPointPackData) initPathBuilders() {
	if !ppd.pathBuildersLoaded {
		err := ppd.loadPathBuilders()
		if err != nil {
			Log.Fatal(err)
			return
		}
		Log.Debugf("Environment %d has %d path builders", ppd.GetEnvId(), len(ppd.pathBuilders))
	}
}

func (ppd *ServerPointPackData) FillDb() {
	n, err := ppd.saveAllConnectionDetails()
	if err != nil {
		Log.Fatalf("could not save all connections due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d connection details", ppd.GetEnvId(), n)
	}

	ppd.initConnections()

	n, err = ppd.saveAllTrioDetails()
	if err != nil {
		Log.Fatalf("could not save all trios due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio details", ppd.GetEnvId(), n)
	}
	ppd.initTrioDetails()

	n, err = ppd.saveAllGrowthContexts()
	if err != nil {
		Log.Fatalf("could not save all growth contexts due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d growth contexts", ppd.GetEnvId(), n)
	}
	ppd.initGrowthContexts()

	n, err = ppd.saveAllContextCubes()
	if err != nil {
		Log.Fatalf("could not save all contexts cubes due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d contexts cubes", ppd.GetEnvId(), n)
	}
	ppd.initContextCubes()

	n, err = ppd.saveAllPathBuilders()
	if err != nil {
		Log.Fatalf("could not save all path builders due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d path builders", ppd.GetEnvId(), n)
	}
	ppd.initPathBuilders()
}
