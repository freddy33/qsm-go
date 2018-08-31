package main

import (
	"os"
	"fmt"
	"github.com/freddy33/qsm-go/spacegl"
	"github.com/freddy33/qsm-go/playgl"
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
	case "cube1":
		spacegl.DisplayCube1()
	case "cube2":
		spacegl.DisplayCube2()
	default:
		fmt.Println("The param",c,"unknown")
	}
	fmt.Println("Finished Executing", c)
}

