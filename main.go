package main

import (
	"os"
	"fmt"
	"github.com/freddy33/qsm-go/playgl"
	"github.com/freddy33/qsm-go/m3space"
)

func main() {
	c := "play1"
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	fmt.Println("Executing", c)
	switch c {
	case "play1":
		playgl.DisplayPlay1()
	case "writeTables":
		m3space.WriteAllTables()
	case "writeTimeFlows":
		m3space.Write0To4TimeFlow()
	case "writeDataFiles":
		m3space.GenerateDataTimeFlow0()
	default:
		fmt.Println("The param",c,"unknown")
	}
	fmt.Println("Finished Executing", c)
}

