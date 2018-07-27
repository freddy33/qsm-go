package space_gl

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"fmt"
	"time"
)

func InitWindow(width, height int, title string) (win *glfw.Window, err error) {
	if err = glfw.Init(); err != nil {
		return
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	//glfw.WindowHint(glfw.Samples, 4)

	win, err = glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return
	}
	win.MakeContextCurrent()
	return win, nil
}

func InitGl(win *glfw.Window) (err error) {
	if err = gl.Init(); err != nil {
		return
	}

	actualWidth, actualHeight := win.GetSize()
	gl.Viewport(0, 0, int32(actualWidth), int32(actualHeight))

	return nil
}

func ScaleView(win *glfw.Window) (width, height int){
	width, height = win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	return
}

func Init3DGl(win *glfw.Window) (prog uint32, err error) {
	if err = gl.Init(); err != nil {
		return
	}
	win.SetKeyCallback(onKey)

	fmt.Println("Renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println("OpenGL version suppported::", gl.GoStr(gl.GetString(gl.VERSION)))
	time.Sleep(time.Microsecond*100)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Disable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	prog = gl.CreateProgram()
	gl.LinkProgram(prog)
	return prog,nil
}

func onKey(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && key == glfw.KeyEscape {
		win.SetShouldClose(true)
	}
}
