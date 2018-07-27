package space_gl

import (
	"log"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/all-core/gl"
)

const windowWidth = 800
const windowHeight = 600

func DisplayCube2() {
	win, err := InitWindow(windowWidth, windowHeight, "basic 3d")
	if err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	prog, err := Init3DGl(win)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(prog)

}
