package main

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
	"os"
)

func main() {
	c := "empty"
	m3util.ReadVerbose()
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
	defer m3util.CloseAll()
	switch c {
	case "gentxt":
		m3point.GenerateTextFilesEnv(m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment))
	case "filldb":
		m3point.FillDbEnv(m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment))
	case "refilldb":
		m3point.ReFillDbEnv(m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment))
	case "perf":
		m3path.RunInsertRandomPoints()
	default:
		fmt.Println("The param", c, "unknown")
		os.Exit(1)
	}
	fmt.Println("Finished Executing", c)
}