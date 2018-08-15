package main

import (
	"github.com/freddy33/qsm-go/space_gl"
	"os"
	"fmt"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	c := "gl_cube2"
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	fmt.Println("Executing", c)
	switch c {
	case "gl_cube1":
		space_gl.DisplayCube()
	case "gl_cube2":
		space_gl.DisplayCube2()
	default:
		fmt.Println("The param",c,"unknown")
	}
	fmt.Println("Finished Executing", c)
}

