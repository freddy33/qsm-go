package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"time"
)

var pointEnv *m3db.QsmEnvironment

var connectionsLoaded bool
var trioDetailsLoaded bool
var trioContextsLoaded bool
var cubesPerContextLoaded bool

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

func checkTrioContextsInitialized() {
	if !trioDetailsLoaded {
		Log.Fatal("Trio contexts should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func checkCubesInitialized() {
	if !cubesPerContextLoaded {
		Log.Fatal("Cubes should have been initialized! Please call m3point.Initialize() method before this!")
	}
}

func Initialize() {
	initConnections()
	initTrioDetails()
	initTrioContexts()
	initContextCubes()
}

func initConnections() {
	if !connectionsLoaded {
		allConnections, allConnectionsByVector = loadConnectionDetails()
		connectionsLoaded = true
	}
}

func initTrioDetails() {
	if !trioDetailsLoaded {
		allTrioDetails = loadTrioDetails()
		trioDetailsLoaded = true
	}
}

func initTrioContexts() {
	if !trioContextsLoaded {
		allTrioContexts = loadTrioContexts()
		trioContextsLoaded = true
	}
}

func initContextCubes() {
	if !cubesPerContextLoaded {
		allCubesPerContext = loadContextCubes()
		cubesPerContextLoaded = true
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

	n, err = saveAllTrioContexts()
	if err != nil {
		Log.Fatalf("could not save all trio contexts due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d trio contexts", env.GetId(), n)
	}
	initTrioContexts()

	n, err = saveAllContextCubes()
	if err != nil {
		Log.Fatalf("could not save all contexts cubes due to %v", err)
		return
	}
	if Log.IsInfo() {
		Log.Infof("Environment %d has %d contexts cubes", env.GetId(), n)
	}
	initContextCubes()
}
