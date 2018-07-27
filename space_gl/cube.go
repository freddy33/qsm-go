package space_gl

// Convert from https://stackoverflow.com/users/44729/genpfault of answer in https://stackoverflow.com/questions/24040982/c-opengl-glfw-drawing-a-simple-cube

import (
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"math"
	"fmt"
	"time"
	"log"
)

const (
	winWidth  = 515
	winHeight = 544
)

var alpha float32
var cubeVertices = []float32{
	-1, -1, -1,
	-1, -1, 1,
	-1, 1, 1,
	-1, 1, -1,
	1, -1, -1,
	1, -1, 1,
	1, 1, 1,
	1, 1, -1,
	-1, -1, -1,
	-1, -1, 1, 1, -1, 1, 1, -1, -1,
	-1, 1, -1, -1, 1, 1, 1, 1, 1, 1, 1, -1,
	-1, -1, -1, -1, 1, -1, 1, 1, -1, 1, -1, -1,
	-1, -1, 1, -1, 1, 1, 1, 1, 1, 1, -1, 1}
var cubeColors = []float32{
	0, 0, 0, 0, 0, 1, 0, 1, 1, 0, 1, 0,
	1, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0,
	0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0,
	0, 1, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0,
	0, 0, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0,
	0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1}

func DisplayCube() {
	win, err := InitWindow(winWidth, winHeight, "Cube")
	if err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()
	prog, err := Init3DGl(win)
	if err != nil {
		log.Fatalln("opengl: initialisation failed:", err)
	}
	displayLoop(win, prog)
}

func drawCube() {
	gl.Rotatef(alpha, 0, 0, 0)
	/* We have a color array and a vertex array */
	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.EnableClientState(gl.COLOR_ARRAY)
	gl.VertexPointer(3, gl.FLOAT, 0, gl.Ptr(cubeVertices))
	gl.ColorPointer(3, gl.FLOAT, 0, gl.Ptr(cubeColors))

	/* Send data : 24 vertices */
	gl.DrawArrays(gl.QUADS, 0, 24)

	/* Cleanup states */
	gl.DisableClientState(gl.COLOR_ARRAY)
	gl.DisableClientState(gl.VERTEX_ARRAY)
	alpha += 1
}

func displayLoop(win *glfw.Window, prog uint32) {
	i := 0
	for !win.ShouldClose() {
		// Scale to window size
		//width, height := ScaleView(win)
		width, height := winWidth, winHeight

		if i < 10 {
			fmt.Println("Starting display", i, width, height)
			time.Sleep(time.Microsecond * 1000)
		}
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.UseProgram(prog)
		// Draw stuff
		gl.ClearColor(0.0, 0.8, 0.3, 1.0)

		gl.MatrixMode(gl.PROJECTION_MATRIX)
		gl.LoadIdentity()

		gluPerspective(60, float64(width)/float64(height), 0.1, 100)

		gl.MatrixMode(gl.MODELVIEW_MATRIX)
		gl.Translatef(0, 0, -5)

		drawCube()

		// Check for any input, or window movement
		glfw.PollEvents()

		// Update Screen
		win.SwapBuffers()
		if i < 10 {
			fmt.Println("Finished display", i)
		}
		i++
	}
}

func gluPerspective(fovY, aspect, zNear, zFar float64) {
	fH := math.Tan(fovY/(360*math.Pi)) * zNear
	fW := fH * aspect
	gl.Frustum(-fW, fW, -fH, fH, zNear, zFar)
}
