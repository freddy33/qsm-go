package main

import (
	"github.com/freddy33/qsm-go/ui/playgl"
	"github.com/freddy33/qsm-go/utils/m3util"
)

func main() {
	m3util.ReadVerbose()
	defer m3util.CloseAll()
	playgl.Play()
}
