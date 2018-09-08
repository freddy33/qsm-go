package playgl

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"github.com/go-gl/gl/v4.1-core/gl"
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

	// Configure the vertex and fragment shaders
	prog, err := newProgram(vertexShaderFull, fragmentShader)
	if err != nil {
		panic(err)
	}

	win.SetKeyCallback(onKey)

	projectionUniform := gl.GetUniformLocation(prog, gl.Str("projection\x00"))
	cameraUniform := gl.GetUniformLocation(prog, gl.Str("camera\x00"))
	modelUniform := gl.GetUniformLocation(prog, gl.Str("model\x00"))
	lightDirectionUniform := gl.GetUniformLocation(prog, gl.Str("light_direction\x00"))
	lightColorUniform := gl.GetUniformLocation(prog, gl.Str("light_color\x00"))
	gl.BindFragDataLocation(prog, 0, gl.Str("out_color\x00"))

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	fmt.Println("Nb vertices", w.NbVertices, ", total size", len(w.OpenGLBuffer))
	gl.BufferData(gl.ARRAY_BUFFER, w.NbVertices*m3gl.FloatPerVertices*m3gl.FloatSize, gl.Ptr(w.OpenGLBuffer), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, m3gl.FloatPerVertices*m3gl.FloatSize, gl.PtrOffset(0))

	normAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("norm\x00")))
	gl.EnableVertexAttribArray(normAttrib)
	gl.VertexAttribPointer(normAttrib, 3, gl.FLOAT, true, m3gl.FloatPerVertices*m3gl.FloatSize, gl.PtrOffset(3*m3gl.FloatSize))

	colorAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("obj_color\x00")))
	gl.EnableVertexAttribArray(colorAttrib)
	gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, m3gl.FloatPerVertices*m3gl.FloatSize, gl.PtrOffset(6*m3gl.FloatSize))

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
		gl.Uniform3f(lightDirectionUniform, w.LightDirection[0], w.LightDirection[1], w.LightDirection[2])
		gl.Uniform3f(lightColorUniform, w.LightColor[0], w.LightColor[1], w.LightColor[2])
		gl.BindVertexArray(vao)

		for _, toDraw := range w.Objects {
			gl.DrawArrays(gl.TRIANGLES, toDraw.OpenGLOffset, toDraw.NbVertices)
		}

		win.SwapBuffers()
		glfw.PollEvents()
	}

}

func onKey(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	reCalc := false
	reFill := false
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			win.SetShouldClose(true)
		case glfw.KeyS:
			w.Rotate = !w.Rotate
		case glfw.KeyZ:
			w.FovAngle.Decrease()
			reCalc = true
		case glfw.KeyX:
			w.FovAngle.Increase()
			reCalc = true
		case glfw.KeyQ:
			w.EyeDist.Increase()
			reCalc = true
		case glfw.KeyW:
			w.EyeDist.Decrease()
			reCalc = true
		case glfw.KeyB:
			m3gl.LineWidth.Increase()
			reCalc = true
			reFill = true
		case glfw.KeyT:
			m3gl.LineWidth.Decrease()
			reCalc = true
			reFill = true
		case glfw.KeyP:
			m3gl.SphereRadius.Increase()
			reCalc = true
			reFill = true
		case glfw.KeyL:
			m3gl.SphereRadius.Decrease()
			reCalc = true
			reFill = true
		}
	}
	if reCalc {
		recalc(reFill)
	} else {
		w.DisplaySettings()
	}
}

func recalc(fill bool) {
	w.DisplaySettings()
	w.SetMatrices()
	if fill {
		w.CreateObjects()
		gl.BufferData(gl.ARRAY_BUFFER, w.NbVertices*m3gl.FloatPerVertices*4, gl.Ptr(w.OpenGLBuffer), gl.STATIC_DRAW)
	}
}

var vertexShaderFull = `
#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 vert;
in vec3 norm;
in vec3 obj_color;

out vec3 s_normal;
out vec3 s_obj_color;

void main() {
	s_normal = vec3(model * vec4(norm, 1));
	s_obj_color = obj_color;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330

uniform vec3 light_direction;
uniform vec3 light_color;

in vec3 s_pos;
in vec3 s_normal;
in vec3 s_obj_color;

out vec4 out_color;

void main() {
    // ambient
    float ambientStrength = 0.15;
    vec3 ambient = ambientStrength * light_color;
  	
    // diffuse 
    float diff = max(dot(s_normal, light_direction), 0.0);
    vec3 diffuse = diff * light_color;
            
    vec3 result = (ambient + diffuse) * s_obj_color;
    out_color = vec4(result, 1.0);
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
