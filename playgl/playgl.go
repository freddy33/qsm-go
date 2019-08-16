package playgl

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3gl"
	"github.com/freddy33/qsm-go/m3path"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3space"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"runtime"
	"strings"
)

var Log = m3util.NewLogger("playgl", m3util.INFO)

const windowWidth = 800
const windowHeight = 600

// TODO: Is there another way than global?
var world m3gl.DisplayWorld

func Play() {
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		Log.Fatalf("could not initialize glfw: %v", err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(windowWidth, windowHeight, "Hello world", nil, nil)
	if err != nil {
		Log.Fatalf("could not create opengl renderer: %v", err)
	}
	defer glfw.Terminate()
	win.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		Log.Fatal(err)
	}

	Log.Info("Renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	Log.Info("OpenGL version supported:", gl.GoStr(gl.GetString(gl.VERSION)))

	// ******************************************************************
	//    HERE CHANGE THE SIZE
	// ******************************************************************
	max := int64(30 * m3point.THREE)
	env := m3db.GetDefaultEnvironment()
	m3path.InitializeDBEnv(env)
	world = m3gl.MakeWorld(env, max, glfw.GetTime())
	//world.WorldSpace.CreateSingleEventCenter()
	world.WorldSpace.EventOutgrowthThreshold = m3space.DistAndTime(0)
	world.WorldSpace.EventOutgrowthOldThreshold = m3space.DistAndTime(50)
	world.WorldSpace.CreatePyramid(20)
	world.CreateDrawingElements()
	world.WorldSpace.MaxConnections = 3

	// Configure the vertex and fragment shaders
	prog, err := newProgram(vertexShaderFull, fragmentShader)
	if err != nil {
		Log.Fatal(err)
	}

	win.SetKeyCallback(onKey)

	projectionUniform := gl.GetUniformLocation(prog, gl.Str("projection\x00"))
	cameraUniform := gl.GetUniformLocation(prog, gl.Str("camera\x00"))
	modelUniform := gl.GetUniformLocation(prog, gl.Str("model\x00"))
	colorUniform := gl.GetUniformLocation(prog, gl.Str("obj_color\x00"))
	colorDimmerUniform := gl.GetUniformLocation(prog, gl.Str("obj_dimmer\x00"))
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
	Log.Info("Nb vertices", world.NbVertices, ", total size", len(world.OpenGLBuffer))
	gl.BufferData(gl.ARRAY_BUFFER, world.NbVertices*m3gl.FloatPerVertices*m3gl.FloatSize, gl.Ptr(world.OpenGLBuffer), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, m3gl.FloatPerVertices*m3gl.FloatSize, gl.PtrOffset(0))

	normAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("norm\x00")))
	gl.EnableVertexAttribArray(normAttrib)
	gl.VertexAttribPointer(normAttrib, 3, gl.FLOAT, true, m3gl.FloatPerVertices*m3gl.FloatSize, gl.PtrOffset(3*m3gl.FloatSize))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	for !win.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// In Retina display retrieving the window size give half of what is needed. Using framebuffer size fix the issue.
		world.Width, world.Height = win.GetFramebufferSize()
		gl.Viewport(0, 0, int32(world.Width), int32(world.Height))
		world.Tick(glfw.GetTime())

		gl.UseProgram(prog)

		gl.UniformMatrix4fv(projectionUniform, 1, false, &(world.Projection[0]))
		gl.UniformMatrix4fv(cameraUniform, 1, false, &(world.Camera[0]))
		gl.UniformMatrix4fv(modelUniform, 1, false, &(world.Model[0]))
		gl.Uniform3f(lightDirectionUniform, world.LightDirection[0], world.LightDirection[1], world.LightDirection[2])
		gl.Uniform3f(lightColorUniform, world.LightColor[0], world.LightColor[1], world.LightColor[2])
		gl.Uniform1i(colorUniform, 0)
		gl.Uniform1f(colorDimmerUniform, 1.0)
		gl.BindVertexArray(vao)

		for _, obj := range world.Elements {
			if obj != nil && obj.Display(world.Filter) {
				toDraw := world.DrawingElementsMap[obj.Key()]
				world.Model = mgl32.HomogRotate3D(float32(world.Angle.Value), mgl32.Vec3{0, 0, 1})
				world.Model = world.Model.Mul4(mgl32.Translate3D(float32(obj.Pos().X()), float32(obj.Pos().Y()), float32(obj.Pos().Z())))
				gl.UniformMatrix4fv(modelUniform, 1, false, &(world.Model[0]))

				gl.Uniform1i(colorUniform, obj.Color(world.Blinker.Value))
				gl.Uniform1f(colorDimmerUniform, obj.Dimmer(world.Blinker.Value))

				gl.DrawArrays(gl.TRIANGLES, toDraw.OpenGLOffset, toDraw.NbVertices)
			}
		}

		win.SwapBuffers()
		glfw.PollEvents()
	}

}

func onKey(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		displaySettings := true
		reCalc := false
		reFill := false

		switch key {
		case glfw.KeyEscape:
			win.SetShouldClose(true)
			displaySettings = false

		case glfw.KeyS:
			world.Angle.Enabled = !world.Angle.Enabled
			displaySettings = false

		case glfw.KeyRight:
			world.WorldSpace.ForwardTime()
			if world.CheckMax() {
				gl.BufferData(gl.ARRAY_BUFFER, world.NbVertices*m3gl.FloatPerVertices*4, gl.Ptr(world.OpenGLBuffer), gl.STATIC_DRAW)
			}
			world.CreateDrawingElements()
			displaySettings = false

		case glfw.KeyN:
			world.Filter.DisplayEmptyNodes = !world.Filter.DisplayEmptyNodes
		case glfw.KeyC:
			world.Filter.DisplayEmptyConnections = !world.Filter.DisplayEmptyConnections

		case glfw.KeyUp:
			world.EventOutgrowthThresholdIncrease()
		case glfw.KeyDown:
			world.EventOutgrowthThresholdDecrease()

		case glfw.KeyU:
			world.Filter.EventOutgrowthColorsIncrease()
		case glfw.KeyI:
			world.Filter.EventOutgrowthColorsDecrease()

		case glfw.Key1:
			world.Filter.ColorMaskSwitch(m3space.RedEvent)
		case glfw.Key2:
			world.Filter.ColorMaskSwitch(m3space.GreenEvent)
		case glfw.Key3:
			world.Filter.ColorMaskSwitch(m3space.BlueEvent)
		case glfw.Key4:
			world.Filter.ColorMaskSwitch(m3space.YellowEvent)

		case glfw.KeyZ:
			world.FovAngle.Decrease()
			reCalc = true
		case glfw.KeyX:
			world.FovAngle.Increase()
			reCalc = true
		case glfw.KeyQ:
			world.EyeDist.Increase()
			reCalc = true
		case glfw.KeyW:
			world.EyeDist.Decrease()
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

		if reCalc {
			recalc(reFill)
		} else {
			if displaySettings {
				world.DisplaySettings()
			}
		}
	}
}

func recalc(fill bool) {
	world.DisplaySettings()
	world.SetMatrices()
	if fill {
		world.CreateDrawingElementsMap()
		gl.BufferData(gl.ARRAY_BUFFER, world.NbVertices*m3gl.FloatPerVertices*4, gl.Ptr(world.OpenGLBuffer), gl.STATIC_DRAW)
	}
}

var vertexShaderFull = `
#version 330

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform int obj_color;
uniform float obj_dimmer;

in vec3 vert;
in vec3 norm;

out vec3 s_normal;
out vec3 s_obj_color;

void main() {
// Be careful last val of vec4 is zero since no translation on normal vector
	s_normal = vec3(model * vec4(norm, 0));

    gl_Position = projection * camera * model * vec4(vert, 1);

	if (obj_color == 0) {
		s_obj_color = vec3(0.25,0.25,0.25) * obj_dimmer;
	} else if (obj_color == 1) {
		s_obj_color = vec3(1.0,0.0,0.0) * obj_dimmer;
	} else if (obj_color == 2) {
		s_obj_color = vec3(0.0,1.0,0.0) * obj_dimmer;
	} else if (obj_color == 4) {
		s_obj_color = vec3(0.0,0.0,1.0) * obj_dimmer;
	} else if (obj_color == 8) {
		s_obj_color = vec3(1.0,1.0,0.0) * obj_dimmer;
	} else {
		s_obj_color = vec3(1.0,1.0,1.0) * obj_dimmer;
	}
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

    vec3 result = (ambient + diffuse) * vec3(s_obj_color);
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
