package space_gl

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"fmt"
)

func init() {
	runtime.LockOSThread()
}

func InitWindow(width, height int, title string) (win *glfw.Window, err error) {
	////////////////////////////////////////////
	if err = glfw.Init(); err != nil {
		return
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)

	win, err = glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return
	}
	win.MakeContextCurrent()
	return win, nil
}

func InitGl(win *glfw.Window) (err error) {
	actualWidth, actualHeight := win.GetSize()
	if err = gl.Init(); err != nil {
		return
	}
	gl.Viewport(0, 0, int32(actualWidth), int32(actualHeight))
	return nil
}

func ScaleView(win *glfw.Window) (width, height int){
	width, height = win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	return
}

func Init3DGl(win *glfw.Window) (err error) {
	win.SetKeyCallback(keyCallback)

	if err = gl.Init(); err != nil {
		return
	}
	fmt.Println("Renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println("OpenGL version suppported::", gl.GoStr(gl.GetString(gl.VERSION)))

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Disable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	return nil
}

func keyCallback(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && key == glfw.KeyEscape {
		win.SetShouldClose(true)
	}
}
