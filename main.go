package main

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/playgl"
	"os"
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
		m3point.WriteAllTables()
	default:
		fmt.Println("The param",c,"unknown")
	}
	fmt.Println("Finished Executing", c)
}

