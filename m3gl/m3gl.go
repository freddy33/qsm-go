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
	coordinates        = 3
	FloatPerVertices   = 9 // PPPNNNCCC
	circlePartsLine    = 16
	trianglesPerLine   = circlePartsLine * 2
	circlePartsSphere  = 32
	nbMiddleCircles    = (circlePartsSphere - 2) / 2
	trianglesPerSphere = nbMiddleCircles * circlePartsSphere * 2
	pointsPerTriangle  = 3
)

// QSM DrawingElementsMap const
const (
	nodes       = 4
	connections = 2
	axes        = 3
)

type World struct {
	Max           int64
	TopCornerDist float64

	NbVertices         int
	OpenGLBuffer       []float32
	DrawingElementsMap map[m3space.ObjectType]OpenGLDrawingElement

	Width, Height  int
	EyeDist        SizeVar
	FovAngle       SizeVar
	LightDirection mgl32.Vec3
	LightColor     mgl32.Vec3
	Projection     mgl32.Mat4
	Camera         mgl32.Mat4
	Model          mgl32.Mat4
	previousTime   float64
	previousArea   int64
	Angle          float64
	Rotate         bool
}

type WorldDrawingElement struct {
	DrawEl *OpenGLDrawingElement
	Transform mgl32.Mat4
}

type OpenGLDrawingElement struct {
	k            m3space.ObjectType
	OpenGLOffset int32
	NbVertices   int32
}

func MakeWorld(Max int64) World {
	verifyData()
	TopCornerDist := math.Sqrt(float64(3.0*Max*Max)) + 1.1
	w := World{
		Max,
		TopCornerDist,
		0,
		make([]float32, 0),
		make(map[m3space.ObjectType]OpenGLDrawingElement),
		800, 600,
		SizeVar{float64(Max), TopCornerDist * 4.0, TopCornerDist},
		SizeVar{10.0, 75.0, 30.0},
		mgl32.Vec3{-1.0, 1.0, 1.0}.Normalize(),
		mgl32.Vec3{1.0, 1.0, 1.0},
		mgl32.Ident4(),
		mgl32.Ident4(),
		mgl32.Ident4(),
		glfw.GetTime(),
		0,
		0.0,
		false,
	}
	w.SetMatrices()
	w.CreateObjects()
	m3space.SpaceObj.CreateStuff(Max)
	return w
}

var LineWidth = SizeVar{0.001, 0.5, 0.04}
var SphereRadius = SizeVar{0.05, 0.5, 0.1}
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

func (w World) DisplaySettings() {
	fmt.Println("========= World Settings =========")
	fmt.Println("Nb DrawingElementsMap", len(w.DrawingElementsMap))
	fmt.Println("Width", w.Width, "Height", w.Height)
	fmt.Println("Line Width [B,T]", LineWidth.Val)
	fmt.Println("Sphere Radius [P,L]", SphereRadius.Val)
	fmt.Println("FOV Angle [Z,X]", w.FovAngle.Val)
	fmt.Println("Eye Dist [Q,W]", w.EyeDist.Val)
}

func (w *World) CreateObjects() int {
	nbTriangles := (axes+connections)*trianglesPerLine + (nodes * trianglesPerSphere)
	if w.NbVertices != nbTriangles*3 {
		w.NbVertices = nbTriangles * 3
		fmt.Println("Creating OpenGL buffer for", nbTriangles, "triangles,", w.NbVertices, "vertices,", w.NbVertices*FloatPerVertices, "buffer size.")
		w.OpenGLBuffer = make([]float32, w.NbVertices*FloatPerVertices)
	}
	triangleFiller := TriangleFiller{make(map[m3space.ObjectType]OpenGLDrawingElement), 0, 0,&(w.OpenGLBuffer)}
	p := m3space.Point{w.Max, 0, 0}
	for axe := int16(0); axe < axes; axe++ {
		triangleFiller.fill(MakeSegment(m3space.Origin, p, m3space.ObjectType(axe)))
	}
	for node := 0; node < nodes; node++ {
		triangleFiller.fill(MakeSphere(m3space.ObjectType(int(m3space.NodeA) + node)))
	}
	triangleFiller.fill(MakeSegment(m3space.Origin, m3space.BasePoints[0], m3space.Connection1))
	triangleFiller.fill(MakeSegment(m3space.BasePoints[0], m3space.BasePoints[1].Add(m3space.Point{0, 3, 0}), m3space.Connection2))

	fmt.Println("Created",len(triangleFiller.objMap),"objects")
	w.DrawingElementsMap = triangleFiller.objMap
	fmt.Println("Saved",len(w.DrawingElementsMap),"objects in world map. Keys are:")
	for _, el := range w.DrawingElementsMap {
		fmt.Println(el.k)
	}

	return nbTriangles
}

type Triangle struct {
	vertices [pointsPerTriangle]mgl64.Vec3
	normal   mgl64.Vec3
	color    mgl64.Vec3
}

var AxeXColor = mgl64.Vec3{0.5, 0.2, 0.2}
var AxeYColor = mgl64.Vec3{0.2, 0.5, 0.2}
var AxeZColor = mgl64.Vec3{0.2, 0.2, 0.5}
var Node0Color = mgl64.Vec3{0.7, 0.7, 0.7}
var NodeAColor = mgl64.Vec3{1.0, 1.0, 0.0}
var NodeBColor = mgl64.Vec3{0.0, 1.0, 1.0}
var NodeCColor = mgl64.Vec3{1.0, 0.0, 1.0}
var Conn1Color = mgl64.Vec3{0.8, 0.0, 0.2}
var Conn2Color = mgl64.Vec3{0.2, 0.0, 0.8}

func squareDist(v mgl64.Vec3) float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

func MakeTriangle(vert [3]mgl64.Vec3, t m3space.ObjectType) Triangle {
	AB := vert[1].Sub(vert[0])
	AC := vert[2].Sub(vert[0])
	norm := AB.Cross(AC).Normalize()
	return MakeTriangleWithNorm(vert, norm, t)
}

func MakeTriangleWithNorm(vert [3]mgl64.Vec3, norm mgl64.Vec3, t m3space.ObjectType) Triangle {
	color := mgl64.Vec3{}
	switch t {
	case m3space.AxeX:
		color = AxeXColor
	case m3space.AxeY:
		color = AxeYColor
	case m3space.AxeZ:
		color = AxeZColor
	case m3space.Node0:
		color = Node0Color
	case m3space.NodeA:
		color = NodeAColor
	case m3space.NodeB:
		color = NodeBColor
	case m3space.NodeC:
		color = NodeCColor
	case m3space.Connection1:
		color = Conn1Color
	case m3space.Connection2:
		color = Conn2Color
	}
	return Triangle{vert, norm, color}
}


func (w *World) ScaleView(win *glfw.Window) int64 {
	// In Retina display retrieving the window size give half of what is needed. Using framebuffer size fix the issue.
	w.Width, w.Height = win.GetFramebufferSize()
	gl.Viewport(0, 0, int32(w.Width), int32(w.Height))
	return int64(w.Width) * int64(w.Height)
}

func (w *World) Tick(win *glfw.Window) {
	// Update
	time := glfw.GetTime()
	elapsed := time - w.previousTime
	w.previousTime = time

	area := w.ScaleView(win)
	if area != w.previousArea {
		w.SetMatrices()
		w.previousArea = area
	}

	if w.Rotate {
		w.Angle += elapsed / 2.0
		w.Model = mgl32.HomogRotate3D(float32(w.Angle), mgl32.Vec3{0, 0, 1})
	}
}

func (w *World) SetMatrices() {
	Eye := mgl32.Vec3{float32(w.EyeDist.Val), float32(w.EyeDist.Val), float32(w.EyeDist.Val),}
	Far := Eye.Len() + float32(w.TopCornerDist)
	w.Projection = mgl32.Perspective(mgl32.DegToRad(float32(w.FovAngle.Val)), float32(w.Width)/float32(w.Height), 1.0, Far)
	w.Camera = mgl32.LookAtV(Eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, 1})
}

type TriangleFiller struct {
	objMap           map[m3space.ObjectType]OpenGLDrawingElement
	verticesOffset int32
	bufferOffset int
	buffer       *[]float32
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
			for coord := 0; coord < coordinates; coord++ {
				(*t.buffer)[t.bufferOffset] = float32(tr.color[coord])
				t.bufferOffset++
			}
		}
	}
}
