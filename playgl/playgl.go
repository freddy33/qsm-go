package playgl

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3gl"
	"strings"
)

const windowWidth = 800
const windowHeight = 600

var w m3gl.World

func DisplayPlay1() {
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("could not initialize glfw: %v", err))
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(windowWidth, windowHeight, "Hello world", nil, nil)
	if err != nil {
		panic(fmt.Errorf("could not create opengl renderer: %v", err))
	}
	defer glfw.Terminate()
	win.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	fmt.Println("Renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println("OpenGL version suppported::", gl.GoStr(gl.GetString(gl.VERSION)))

	w = m3gl.MakeWorld(9)
	w.FillAllVertices()

	// Configure the vertex and fragment shaders
	prog, err := newProgram(vertexShaderFull, fragmentShader)
	if err != nil {
		panic(err)
	}

	win.SetKeyCallback(onKey)

	projectionUniform := gl.GetUniformLocation(prog, gl.Str("projection\x00"))
	cameraUniform := gl.GetUniformLocation(prog, gl.Str("camera\x00"))
	modelUniform := gl.GetUniformLocation(prog, gl.Str("model\x00"))
	gl.BindFragDataLocation(prog, 0, gl.Str("out_color\x00"))

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	fmt.Println("Nb vertices", w.NbVertices)
	gl.BufferData(gl.ARRAY_BUFFER, w.NbVertices*3*4+w.NbVertices*2, nil, gl.STATIC_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, w.NbVertices*3*4, gl.Ptr(w.AllVertices))
	gl.BufferSubData(gl.ARRAY_BUFFER, w.NbVertices*3*4, w.NbVertices*2, gl.Ptr(w.AllTypes))

	vertAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	objTypeAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("obj_type\x00")))
	gl.EnableVertexAttribArray(objTypeAttrib)
	gl.VertexAttribIPointer(objTypeAttrib, 1, gl.SHORT, 0, gl.PtrOffset(w.NbVertices*3*4))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	for !win.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		w.Tick(win)

		gl.UseProgram(prog)

		gl.UniformMatrix4fv(projectionUniform, 1, false, &(w.Projection[0]))
		gl.UniformMatrix4fv(cameraUniform, 1, false, &(w.Camera[0]))
		gl.UniformMatrix4fv(modelUniform, 1, false, &(w.Model[0]))
		gl.BindVertexArray(vao)

		gl.DrawArrays(gl.TRIANGLES, 0, int32(w.NbVertices))

		win.SwapBuffers()
		glfw.PollEvents()
	}

}

func onKey(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			win.SetShouldClose(true)
		case glfw.KeyS:
			w.Rotate = !w.Rotate
		case glfw.KeyZ:
			w.FovAngle -= 1.0
			fmt.Println("New FOV Angle", w.FovAngle)
			w.SetMatrices()
		case glfw.KeyX:
			w.FovAngle += 1.0
			fmt.Println("New FOV Angle", w.FovAngle)
			w.SetMatrices()
		case glfw.KeyQ:
			w.Eye = w.Eye.Sub(mgl32.Vec3{1.0, 1.0, 1.0})
			fmt.Println("New Eye", w.Eye)
			w.SetMatrices()
		case glfw.KeyW:
			w.Eye = w.Eye.Add(mgl32.Vec3{1.0, 1.0, 1.0})
			fmt.Println("New Eye", w.Eye)
			w.SetMatrices()
		case glfw.KeyB:
			m3gl.LineWidth += 0.02
			fmt.Println("New Line Width", m3gl.LineWidth)
			w.FillAllVertices()
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, w.NbVertices*3*4, gl.Ptr(w.AllVertices))
		case glfw.KeyT:
			m3gl.LineWidth -= 0.02
			fmt.Println("New Line Width", m3gl.LineWidth)
			w.FillAllVertices()
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, w.NbVertices*3*4, gl.Ptr(w.AllVertices))
		case glfw.KeyP:
			m3gl.SphereSize += 0.02
			fmt.Println("New Sphere size", m3gl.SphereSize)
			w.FillAllVertices()
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, w.NbVertices*3*4, gl.Ptr(w.AllVertices))
		case glfw.KeyL:
			m3gl.SphereSize -= 0.02
			fmt.Println("New Sphere size", m3gl.SphereSize)
			w.FillAllVertices()
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, w.NbVertices*3*4, gl.Ptr(w.AllVertices))
		}
	}
}

var vertexShaderDirect = `
#version 330

uniform mat4 model;

in vec3 vert;

void main() {
    gl_Position = model * vec4(vert, 1);
}
` + "\x00"

var vertexShaderFull = `
#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 vert;
in int obj_type;

out vec3 obj_type_from_shader;

void main() {
    gl_Position = projection * camera * model * vec4(vert, 1);
	if (obj_type == 0) {
		obj_type_from_shader = vec3(0.8,0.0,0.0);
	} else if (obj_type == 1) {
		obj_type_from_shader = vec3(0.0,0.8,0.0);
	} else if (obj_type == 2) {
		obj_type_from_shader = vec3(0.0,0.0,0.8);
	} else if (obj_type == 3) {
		obj_type_from_shader = vec3(1.0,0.3,0.3);
	} else if (obj_type == 4) {
		obj_type_from_shader = vec3(0.3,1.0,0.3);
	} else if (obj_type == 5) {
		obj_type_from_shader = vec3(0.3,0.3,1.0);
	} else if (obj_type == 6) {
		obj_type_from_shader = vec3(0.3,1.0,1.0);
	} else if (obj_type == 7) {
		obj_type_from_shader = vec3(1.0,0.3,1.0);
	} else if (obj_type == 8) {
		obj_type_from_shader = vec3(1.0,1.0,0.3);
	} else {
		obj_type_from_shader = vec3(1.0,1.0,1.0);
	}
}
` + "\x00"

var fragmentShader = `
#version 330

in vec3 obj_type_from_shader;

out vec4 out_color;

void main() {
	out_color = vec4(obj_type_from_shader, 1.0);
}
` + "\x00"

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		logger := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(logger))

		return 0, fmt.Errorf("failed to compile %v: %v", source, logger)
	}

	return shader, nil
}

var demoVertices = []float32{
	0.0, 0.0, 0.0,
	0.5, 0.0, 0.0,
	0.5, 0.5, 0.0,

	0.0, 0.0, 0.0,
	0.0, 0.5, 0.5,
	-0.5, 0.5, 0.5,

	0.0, 0.0, -0.5,
	-0.5, 0.0, 0.5,
	-0.5, -0.5, 0.5,

	0.0, 0.0, 0.0,
	0.0, -0.5, -0.5,
	0.5, -0.5, -0.5,
}
