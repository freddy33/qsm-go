package space_gl

// Convert from https://stackoverflow.com/users/44729/genpfault of answer in https://stackoverflow.com/questions/24040982/c-opengl-glfw-drawing-a-simple-cube

import (
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"math"
	"fmt"
	"time"
	"log"
	"runtime"
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

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func DisplayCube() {
	if err := glfw.Init(); err != nil {
		return
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	win, err := glfw.CreateWindow(winWidth, winHeight, "Cube 1", nil, nil)
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	if err = gl.Init(); err != nil {
		panic(err)
	}

	var prog uint32 = 0
	//prog := gl.CreateProgram()
	//gl.LinkProgram(prog)

	win.SetKeyCallback(onKey)

	fmt.Println("Renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println("OpenGL version suppported::", gl.GoStr(gl.GetString(gl.VERSION)))
	time.Sleep(time.Microsecond*100)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Disable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

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
	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices)*4, gl.Ptr(cubeVertices), gl.STATIC_DRAW)

	for !win.ShouldClose() {
		// Scale to window size
		width, height := ScaleView(win)

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		//gl.UseProgram(prog)
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
	}
}

func gluPerspective(fovY, aspect, zNear, zFar float64) {
	fH := math.Tan(fovY * math.Pi / 360) * zNear
	fW := fH * aspect
	gl.Frustum(-fW, fW, -fH, fH, zNear, zFar)
}

func onKey(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && key == glfw.KeyEscape {
		win.SetShouldClose(true)
	}
}

func ScaleView(win *glfw.Window) (width, height int){
	width, height = win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	return
}
