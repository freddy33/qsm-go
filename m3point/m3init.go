package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"time"
)

var pointEnv *m3db.QsmEnvironment

var connectionsLoaded bool
var trioDetailsLoaded bool
var growthContextsLoaded bool
var cubesLoaded bool
var pathBuildersLoaded bool

func checkConnInitialized() {
	if !connectionsLoaded {
		Log.Fatal("Connections should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkTrioInitialized() {
	if !trioDetailsLoaded {
		Log.Fatal("Trios should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkGrowthContextsInitialized() {
	if !trioDetailsLoaded {
		Log.Fatal("trio contexts should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkCubesInitialized() {
	if !cubesLoaded {
		Log.Fatal("Cubes should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkPathBuildersInitialized() {
	if !pathBuildersLoaded {
		Log.Fatal("Path Builders should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func Initialize() {
	initConnections()
	initTrioDetails()
	initGrowthContexts()
	initContextCubes()
	initPathBuilders()
}

func initConnections() {
	if !connectionsLoaded {
		allConnections, allConnectionsByVector = loadConnectionDetails()
		connectionsLoaded = true
		Log.Debugf("Environment %d has %d connection details", GetPointEnv().GetId(), len(allConnections))
	}
}

func initTrioDetails() {
	if !trioDetailsLoaded {
		allTrioDetails = loadTrioDetails()
		trioDetailsLoaded = true
		Log.Debugf("Environment %d has %d trio details", GetPointEnv().GetId(), len(allTrioDetails))
	}
}

func initGrowthContexts() {
	if !growthContextsLoaded {
		allGrowthContexts = loadGrowthContexts()
		growthContextsLoaded = true
		Log.Debugf("Environment %d has %d growth contexts", GetPointEnv().GetId(), len(allGrowthContexts))
	}
}

func initContextCubes() {
	if !cubesLoaded {
		cubeIdsPerKey = loadContextCubes()
		cubesLoaded = true
		Log.Debugf("Environment %d has %d cubes", GetPointEnv().GetId(), len(cubeIdsPerKey))
	}
}

func initPathBuilders() {
	if !pathBuildersLoaded {
		pathBuilders = loadPathBuilders()
		pathBuildersLoaded = true
		Log.Debugf("Environment %d has %d path builders", GetPointEnv().GetId(), len(pathBuilders))
	}
}

func GetPointEnv() *m3db.QsmEnvironment {
	if pointEnv == nil || pointEnv.GetConnection() == nil {
		pointEnv = m3db.GetDefaultEnvironment()
	}
	return pointEnv
}

func ReFillDb() {
	env := GetPointEnv()
	env.Destroy()
	pointEnv = nil
	time.Sleep(100 * time.Millisecond)
	FillDb()
}

func FillDb() {
	env := GetPointEnv()

	n, err := saveAllConnectionDetails()
	if err != nil {
		Log.Fatalf("could not save all connections due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d connection details", env.GetId(), n)
	}

	initConnections()

	n, err = saveAllTrioDetails()
	if err != nil {
		Log.Fatalf("could not save all trios due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio details", env.GetId(), n)
	}
	initTrioDetails()

	n, err = saveAllGrowthContexts()
	if err != nil {
		Log.Fatalf("could not save all trio contexts due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio contexts", env.GetId(), n)
	}
	initGrowthContexts()

	n, err = saveAllContextCubes()
	if err != nil {
		Log.Fatalf("could not save all contexts cubes due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d contexts cubes", env.GetId(), n)
	}
	initContextCubes()

	n, err = saveAllPathBuilders()
	if err != nil {
		Log.Fatalf("could not save all path builders due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d path builders", env.GetId(), n)
	}
	initPathBuilders()
}
