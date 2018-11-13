package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/v4.1-core/gl"
	"math"
	"github.com/go-gl/mathgl/mgl64"
)

// OpenGL const
const (
	FloatSize          = 4
	IntSize            = 2
	coordinates        = 3
	FloatPerVertices   = 6 // PPPNNN
	IntPerVertices     = 1 // C
	circlePartsLine    = 16
	trianglesPerLine   = circlePartsLine * 2
	circlePartsSphere  = 32
	nbMiddleCircles    = (circlePartsSphere - 2) / 2
	trianglesPerSphere = nbMiddleCircles * circlePartsSphere * 2
	pointsPerTriangle  = 3
)

// QSM DrawingElementsMap const
const (
	nodes       = 2
	connections = 25
	axes        = 3
)

type DisplayWorld struct {
	Space  m3space.Space
	Filter m3space.SpaceDrawingFilter

	NbVertices         int
	OpenGLBuffer       []float32
	DrawingElementsMap map[m3space.ObjectType]OpenGLDrawingElement

	TopCornerDist  float64
	Width, Height  int
	EyeDist        SizeVar
	FovAngle       SizeVar
	LightDirection mgl32.Vec3
	LightColor     mgl32.Vec3
	Projection     mgl32.Mat4
	Camera         mgl32.Mat4
	Model          mgl32.Mat4
	previousArea   int64
	Angle          TimeAutoVar
	Blinker        TimeAutoVar
}

type TimeAutoVar struct {
	Enabled      bool
	Threshold    float64
	Ratio        float64
	previousTime float64
	Value        float64
}

type OpenGLDrawingElement struct {
	k            m3space.ObjectType
	OpenGLOffset int32
	NbVertices   int32
}

func MakeWorld(Max int64) DisplayWorld {
	if Max%m3space.THREE != 0 {
		panic(fmt.Sprintf("cannot have a max %d not dividable by %d", Max, m3space.THREE))
	}
	verifyData()
	TopCornerDist := math.Sqrt(float64(3.0*Max*Max)) + 1.1
	w := DisplayWorld{
		m3space.MakeSpace(Max),
		m3space.SpaceDrawingFilter{false, false, uint8(0xFF), 0, nil,},
		0,
		make([]float32, 0),
		make(map[m3space.ObjectType]OpenGLDrawingElement),
		TopCornerDist,
		800, 600,
		SizeVar{float64(Max), TopCornerDist * 4.0, TopCornerDist * 1.5},
		SizeVar{10.0, 75.0, 30.0},
		mgl32.Vec3{-1.0, 1.0, 1.0}.Normalize(),
		mgl32.Vec3{1.0, 1.0, 1.0},
		mgl32.Ident4(),
		mgl32.Ident4(),
		mgl32.Ident4(),
		0,
		TimeAutoVar{false, 0.01, 0.3, glfw.GetTime(), 0.0,},
		TimeAutoVar{true, 0.5, 2.0, glfw.GetTime(), 0.0,},
	}
	w.Filter.Space = &(w.Space)
	w.SetMatrices()
	w.CreateObjects()
	return w
}

var LineWidth = SizeVar{0.05, 0.5, 0.18}
var SphereRadius = SizeVar{0.1, 0.8, 0.5}
var XH = mgl64.Vec3{1.0, 0.0, 0.0}
var YH = mgl64.Vec3{0.0, 1.0, 0.0}
var ZH = mgl64.Vec3{0.0, 0.0, 1.0}
var XYZ = [3]mgl64.Vec3{XH, YH, ZH}
var CircleForLine = make([]mgl64.Vec2, circlePartsLine)
var CircleForSphere = make([]mgl64.Vec2, circlePartsSphere)

func verifyData() {
	// Verify we capture the equator
	if nbMiddleCircles%2 == 0 {
		panic(fmt.Errorf("something fishy with circle parts %d since %d should be odd", circlePartsSphere, nbMiddleCircles))
	}
	deltaAngle := 2.0 * math.Pi / circlePartsLine
	angle := 0.0
	for i := 0; i < circlePartsLine; i++ {
		CircleForLine[i] = mgl64.Vec2{math.Cos(angle), math.Sin(angle)}
		angle += deltaAngle
	}
	deltaAngle = 2.0 * math.Pi / circlePartsSphere
	angle = 0.0
	for i := 0; i < circlePartsSphere; i++ {
		CircleForSphere[i] = mgl64.Vec2{math.Cos(angle), math.Sin(angle)}
		angle += deltaAngle
	}
}

func (world DisplayWorld) DisplaySettings() {
	fmt.Println("========= DisplayWorld Settings =========")
	fmt.Println("Width", world.Width, "Height", world.Height)
	fmt.Println("Line Width [B,T]", LineWidth.Val)
	fmt.Println("Sphere Radius [P,L]", SphereRadius.Val)
	fmt.Println("FOV Angle [Z,X]", world.FovAngle.Val)
	fmt.Println("Eye Dist [Q,W]", world.EyeDist.Val)
	world.Space.DisplaySettings()
	world.Filter.DisplaySettings()
}

func (world *DisplayWorld) CreateObjects() int {
	m3space.InitConnectionDetails()
	nbTriangles := (axes+connections)*trianglesPerLine + (nodes * trianglesPerSphere)
	if world.NbVertices != nbTriangles*3 {
		world.NbVertices = nbTriangles * 3
		fmt.Println("Creating OpenGL buffer for", nbTriangles, "triangles,", world.NbVertices, "vertices,", world.NbVertices*FloatPerVertices, "buffer size.")
		world.OpenGLBuffer = make([]float32, world.NbVertices*FloatPerVertices)
	}
	triangleFiller := TriangleFiller{make(map[m3space.ObjectType]OpenGLDrawingElement), 0, 0, &(world.OpenGLBuffer)}
	for axe := int16(0); axe < axes; axe++ {
		p := m3space.Point{}
		p[axe] = world.Space.Max + m3space.AxeExtraLength
		triangleFiller.fill(MakeSegment(m3space.Origin, p, m3space.ObjectType(axe)))
	}
	triangleFiller.fill(MakeSphere(m3space.NodeEmpty))
	triangleFiller.fill(MakeSphere(m3space.NodeActive))
	connDone := make(map[uint8]bool)
	for _, cd := range m3space.AllConnectionsPossible {
		if !cd.ConnNeg && !connDone[cd.ConnNumber] {
			triangleFiller.fill(MakeSegment(m3space.Origin, cd.Vector, m3space.ObjectType(uint8(m3space.Connection00)+cd.ConnNumber)))
			connDone[cd.ConnNumber] = true
		}
	}

	world.DrawingElementsMap = triangleFiller.objMap
	fmt.Println("Saved", len(world.DrawingElementsMap), "objects in world map.")

	return nbTriangles
}

type Triangle struct {
	vertices [pointsPerTriangle]mgl64.Vec3
	normal   mgl64.Vec3
}

func MakeTriangle(vert [3]mgl64.Vec3) Triangle {
	AB := vert[1].Sub(vert[0])
	AC := vert[2].Sub(vert[0])
	norm := AB.Cross(AC).Normalize()
	return MakeTriangleWithNorm(vert, norm)
}

func MakeTriangleWithNorm(vert [3]mgl64.Vec3, norm mgl64.Vec3) Triangle {
	return Triangle{vert, norm}
}

func (world *DisplayWorld) ScaleView(win *glfw.Window) int64 {
	// In Retina display retrieving the window size give half of what is needed. Using framebuffer size fix the issue.
	world.Width, world.Height = win.GetFramebufferSize()
	gl.Viewport(0, 0, int32(world.Width), int32(world.Height))
	return int64(world.Width) * int64(world.Height)
}

func (t *TimeAutoVar) Tick(win *glfw.Window) {
	time := glfw.GetTime()
	if t.Enabled {
		elapsed := time - t.previousTime
		if elapsed > t.Threshold {
			t.previousTime = time
			t.Value += elapsed * t.Ratio
		}
	} else {
		// No change just previous time
		t.previousTime = time
	}
}

func (world *DisplayWorld) Tick(win *glfw.Window) {
	area := world.ScaleView(win)
	if area != world.previousArea {
		world.SetMatrices()
		world.previousArea = area
	}
	world.Angle.Tick(win)
	world.Blinker.Tick(win)
	if int32(world.Blinker.Value) >= 4 {
		world.Blinker.Value = 0.0
	}
}

func (world *DisplayWorld) SetMatrices() {
	Eye := mgl32.Vec3{float32(world.EyeDist.Val), float32(world.EyeDist.Val), float32(world.EyeDist.Val),}
	Far := Eye.Len() + float32(world.TopCornerDist)
	world.Projection = mgl32.Perspective(mgl32.DegToRad(float32(world.FovAngle.Val)), float32(world.Width)/float32(world.Height), 1.0, Far)
	world.Camera = mgl32.LookAtV(Eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, 1})
}

type TriangleFiller struct {
	objMap         map[m3space.ObjectType]OpenGLDrawingElement
	verticesOffset int32
	bufferOffset   int
	buffer         *[]float32
}

func (t *TriangleFiller) fill(o GLObject) {
	wo := OpenGLDrawingElement{
		o.Key(),
		t.verticesOffset,
		int32(o.NumberOfVertices()),
	}
	t.objMap[wo.k] = wo
	triangles := o.ExtractTriangles()
	for _, tr := range triangles {
		for _, point := range tr.vertices {
			t.verticesOffset++
			for coord := 0; coord < coordinates; coord++ {
				(*t.buffer)[t.bufferOffset] = float32(point[coord])
				t.bufferOffset++
			}
			for coord := 0; coord < coordinates; coord++ {
				(*t.buffer)[t.bufferOffset] = float32(tr.normal[coord])
				t.bufferOffset++
			}
		}
	}
}
