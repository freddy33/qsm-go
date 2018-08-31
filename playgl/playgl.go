package playgl

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3gl"
	"github.com/freddy33/qsm-go/m3space"
	"math"
	"strings"
)

const windowWidth = 800
const windowHeight = 600

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

	w := makeWorld(1)
	w.createAxes()

	// Configure the vertex and fragment shaders
	prog, err := newProgram(vertexShaderFull, fragmentShader)
	if err != nil {
		panic(err)
	}

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
	gl.BufferData(gl.ARRAY_BUFFER, len(w.AxesVertices)*4, gl.Ptr(w.AxesVertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(prog, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.PROGRAM_POINT_SIZE)
	gl.PointSize(10.0)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	for !win.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		w.tick()

		gl.UseProgram(prog)

		gl.UniformMatrix4fv(projectionUniform, 1, false, &(w.Projection[0]))
		gl.UniformMatrix4fv(cameraUniform, 1, false, &(w.Camera[0]))
		gl.UniformMatrix4fv(modelUniform, 1, false, &(w.Model[0]))
		gl.BindVertexArray(vao)

		gl.DrawArrays(gl.TRIANGLES, 0, 18)
		//gl.DrawArrays(gl.POINTS, 0, 12)

		win.SwapBuffers()
		glfw.PollEvents()
	}

}

const (
	axes              = 3
	trianglePerLine   = 2
	pointsPerTriangle = 3
	coordinates       = 3
)

type World struct {
	Max                       int64
	AxesVertices              []float32
	Eye                       mgl32.Vec3
	Projection, Camera, Model mgl32.Mat4
	previousTime              float64
	angle                     float32
}

func makeWorld(Max int64) World {
	eyeFromOrig := float32(math.Sqrt(float64(3.0*Max*Max)) * 1.3)
	far := eyeFromOrig * 2.0
	eye := mgl32.Vec3{eyeFromOrig, eyeFromOrig, eyeFromOrig}
	w := World{
		Max,
		make([]float32, axes*trianglePerLine*pointsPerTriangle*coordinates),
		eye,
		mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 1.0, far),
		mgl32.LookAtV(eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0}),
		mgl32.Ident4(),
		glfw.GetTime(),
		0.0,
	}
	return w
}

func (w *World) tick() {
	// Update
	time := glfw.GetTime()
	elapsed := time - w.previousTime
	w.previousTime = time

	w.angle += float32(elapsed / 2.0)
	w.Model = mgl32.HomogRotate3D(w.angle, mgl32.Vec3{0, 1, 0})
}

func (w *World) createAxes() {
	offset := 0
	for axe := 0; axe < axes; axe++ {
		p1 := m3space.Point{0, 0, 0}
		p2 := m3space.Point{0, 0, 0}
		p1[axe] = -w.Max
		p2[axe] = w.Max
		axis := m3gl.MakeSegment(p1, p2)
		triangles := axis.ExtractTriangles(w.Eye)
		if len(triangles) != trianglePerLine {
			panic(fmt.Sprint("Number of triangles per lines inconsistent", len(triangles), trianglePerLine))
		}
		for _, triangle := range triangles {
			for _, point := range triangle.Points {
				for coord := 0; coord < coordinates; coord++ {
					w.AxesVertices[offset] = point[coord]
					offset++
				}
			}
		}
	}
}

func (w *World) half() {
	for i, c := range w.AxesVertices {
		w.AxesVertices[i] = c / 2.0
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

void main() {
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var fragmentShader = `
#version 330

out vec4 out_color;

void main() {
    out_color = vec4(1.0, 1.0, 1.0, 1.0);
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
