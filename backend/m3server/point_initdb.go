package m3server

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3util"
)

type PointPackData struct {
	m3point.BasePointPackData
	Env *m3db.QsmDbEnvironment
}

func getServerPointPackData(env m3util.QsmEnvironment) (*PointPackData, bool) {
	newData := env.GetData(m3util.PointIdx) == nil
	if newData {
		ppd := new(PointPackData)
		ppd.EnvId = env.GetId()
		ppd.Env = env.(*m3db.QsmDbEnvironment)
		env.SetData(m3util.PointIdx, ppd)
		// do not return ppd but always the pointer in env data array
	}
	return env.GetData(m3util.PointIdx).(*PointPackData), newData
}

func InitializePointDBEnv(env *m3db.QsmDbEnvironment, forced bool) {
	ppd, newData := getServerPointPackData(env)
	if forced {
		ppd.ResetFlags()
	} else if !newData {
		return
	}
	ppd.initConnections()
	ppd.initTrioDetails()
	ppd.initGrowthContexts()
	ppd.initContextCubes()
	ppd.initPathBuilders()
}

func (ppd *PointPackData) GetValidNextTrio() [12][2]m3point.TrioIndex {
	return validNextTrio
}

func (ppd *PointPackData) GetAllMod4Permutations() [12][4]m3point.TrioIndex {
	return AllMod4Permutations
}

func (ppd *PointPackData) GetAllMod8Permutations() [12][8]m3point.TrioIndex {
	return AllMod8Permutations
}

func (ppd *PointPackData) initConnections() {
	if !ppd.ConnectionsLoaded {
		ppd.AllConnections, ppd.AllConnectionsByVector = loadConnectionDetails(ppd.Env)
		ppd.ConnectionsLoaded = true
		Log.Debugf("Environment %d has %d connection details", ppd.GetEnvId(), len(ppd.AllConnections))
	}
}

func (ppd *PointPackData) initTrioDetails() {
	if !ppd.TrioDetailsLoaded {
		ppd.AllTrioDetails = ppd.loadTrioDetails()
		ppd.TrioDetailsLoaded = true
		Log.Debugf("Environment %d has %d trio details", ppd.GetEnvId(), len(ppd.AllTrioDetails))
	}
}

func (ppd *PointPackData) initGrowthContexts() {
	if !ppd.GrowthContextsLoaded {
		ppd.AllGrowthContexts = ppd.loadGrowthContexts()
		ppd.GrowthContextsLoaded = true
		Log.Debugf("Environment %d has %d growth contexts", ppd.GetEnvId(), len(ppd.AllGrowthContexts))
	}
}

func (ppd *PointPackData) initContextCubes() {
	if !ppd.CubesLoaded {
		ppd.CubeIdsPerKey = ppd.loadContextCubes()
		ppd.CubesLoaded = true
		Log.Debugf("Environment %d has %d cubes", ppd.GetEnvId(), len(ppd.CubeIdsPerKey))
	}
}

func (ppd *PointPackData) initPathBuilders() {
	if !ppd.PathBuildersLoaded {
		ppd.PathBuilders = ppd.loadPathBuilders()
		ppd.PathBuildersLoaded = true
		Log.Debugf("Environment %d has %d path builders", ppd.GetEnvId(), len(ppd.PathBuilders))
	}
}

func FillDbEnv(env *m3db.QsmDbEnvironment) {
	ppd, _ := getServerPointPackData(env)

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
