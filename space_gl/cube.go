package space_gl

import (
	"github.com/xlab/closer"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"math"
)

const (
	winWidth         = 515
	winHeight        = 544
)

var alpha float32
var cubeVertices = []float32{
	-1, -1, -1, -1, -1, 1, -1, 1, 1, -1, 1, -1,
	1, -1, -1, 1, -1, 1, 1, 1, 1, 1, 1, -1,
	-1, -1, -1, -1, -1, 1, 1, -1, 1, 1, -1, -1,
	-1, 1, -1, -1, 1, 1, 1, 1, 1, 1, 1, -1,
	-1, -1, -1, -1, 1, -1, 1, 1, -1, 1, -1, -1,
	-1, -1, 1, -1, 1, 1, 1, 1, 1, 1, -1, 1}
var cubeColors = []float32{
	0, 0, 0,   0, 0, 1,   0, 1, 1,   0, 1, 0,
	1, 0, 0,   1, 0, 1,   1, 1, 1,   1, 1, 0,
	0, 0, 0,   0, 0, 1,   1, 0, 1,   1, 0, 0,
	0, 1, 0,   0, 1, 1,   1, 1, 1,   1, 1, 0,
	0, 0, 0,   0, 1, 0,   1, 1, 0,   1, 0, 0,
	0, 0, 1,   0, 1, 1,   1, 1, 1,   1, 0, 1}

func DisplayCube() {
	win, err := InitWindow(winWidth, winHeight, "Cube")
	if err != nil {
		closer.Fatalln(err)
	}
	err = Init3DGl(win)
	if err != nil {
		closer.Fatalln("opengl: initialisation failed:", err)
	}
	displayLoop(win)
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

func displayLoop(win *glfw.Window) {
	for !win.ShouldClose() {
		// Scale to window size
		width, height := ScaleView(win)

		// Draw stuff
		gl.ClearColor(0.0, 0.8, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.MatrixMode(gl.PROJECTION_MATRIX)
		gl.LoadIdentity()

		gluPerspective( 60, float64(width) / float64(height), 0.1, 100 )

		gl.MatrixMode(gl.MODELVIEW_MATRIX)
		gl.Translatef(0,0,-5)

		drawCube()

		// Update Screen
		win.SwapBuffers()

		// Check for any input, or window movement
		glfw.PollEvents()
	}
}

func gluPerspective(fovY, aspect, zNear, zFar float64) {
	fH := math.Tan( fovY / (360 * math.Pi)) * zNear
	fW := fH * aspect
	gl.Frustum(-fW, fW, -fH, fH, zNear, zFar)
}
