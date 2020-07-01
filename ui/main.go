package main

import (
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
	"github.com/freddy33/qsm-go/ui/m3gl"
	"github.com/freddy33/qsm-go/ui/playgl"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
	"os"
)

func main() {
	m3util.ReadVerbose()
	if len(os.Args) > 1 {
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
	defer m3util.CloseAll()
	playgl.Play()
}
