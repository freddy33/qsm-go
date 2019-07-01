package main

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3gl"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3space"
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
				m3space.Log.SetDebug()
				m3gl.Log.SetDebug()
			}
		}
	}
	fmt.Println("Executing", c)
	defer m3db.CloseAll()
	switch c {
	case "play":
		playgl.Play()
	case "gentxt":
		m3point.GenerateTextFiles()
	case "filldb":
		m3point.FillDb()
	default:
		fmt.Println("The param", c, "unknown")
	}
	fmt.Println("Finished Executing", c)
}
