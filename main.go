package main

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/playgl"
	"os"
)

func main() {
	c := "play"
	if len(os.Args) > 1 {
		c = os.Args[1]
		if len(os.Args) > 2 {
			if os.Args[2] == "-v" {
				m3util.Log.SetDebug()
				m3db.Log.SetDebug()
				m3point.Log.SetDebug()
				m3path.Log.SetDebug()
				//m3space.Log.SetDebug()
				//m3gl.Log.SetDebug()
			}
		}
	}
	fmt.Println("Executing", c)
	defer m3db.CloseAll()
	switch c {
	case "play":
		//fmt.Println("Not yet ready to play full DB mode")
		playgl.Play()
	case "gentxt":
		m3point.GenerateTextFilesEnv(m3db.GetDefaultEnvironment())
	case "filldb":
		m3point.FillDbEnv(m3db.GetDefaultEnvironment())
	case "refilldb":
		m3point.ReFillDbEnv(m3db.GetDefaultEnvironment())
	case "perf":
		m3path.RunInsertRandomPoints()
	default:
		fmt.Println("The param", c, "unknown")
		os.Exit(1)
	}
	fmt.Println("Finished Executing", c)
}
