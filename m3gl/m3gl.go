package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/v4.1-core/gl"
	"math"
)

const (
	FloatPerVertices = 9

	nodes              = 4
	connections        = 3
	axes               = 3
	circleParts        = 8
	trianglesPerLine   = circleParts * 2
	trianglesPerSphere = circleParts * (circleParts - 2)
	pointsPerTriangle  = 3
	coordinates        = 3
)

type ObjectType int16

const (
	axeX       ObjectType = iota
	axeY
	axeZ
	nodeA
	nodeB
	nodeC
	connection
)

type World struct {
	Max                  int64
	TopCornerDist        float32
	NbTriangles          int
	AxesTriangles        []Triangle
	NodesTriangles       []Triangle
	ConnectionsTriangles []Triangle
	NbVertices           int
	OpenGLBuffer         []float32
	Width, Height        int
	EyeDist              SizeVar
	FovAngle             SizeVar
	LightDirection       mgl32.Vec3
	LightColor           mgl32.Vec3
	Projection           mgl32.Mat4
	Camera               mgl32.Mat4
	Model                mgl32.Mat4
	previousTime         float64
	previousArea         int64
	Angle                float32
	Rotate               bool
}

func MakeWorld(Max int64) World {
	TopCornerDist := float32(math.Sqrt(float64(3.0*Max*Max))) + 1.1
	w := World{
		Max,
		TopCornerDist,
		0,
		make([]Triangle, 2*axes*trianglesPerLine),
		make([]Triangle, nodes*trianglesPerSphere),
		make([]Triangle, connections*trianglesPerLine),
		0,
		make([]float32, 0),
		800, 600,
		SizeVar{float32(Max), TopCornerDist * 4.0, TopCornerDist},
		SizeVar{10.0, 75.0, 30.0},
		mgl32.Vec3{1.0, 0.0, 1.0}.Normalize(),
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
	return w
}

var LineWidth = SizeVar{0.001, 0.5, 0.04}
var SphereRadius = SizeVar{0.05, 0.5, 0.1}

func (w World) DisplaySettings() {
	fmt.Println("========= World Settings =========")
	fmt.Println("Axe X", axeX)
	fmt.Println("Connection", connection)
	fmt.Println("Line Width", LineWidth.Val)
	fmt.Println("Sphere Radius", SphereRadius.Val)
	fmt.Println("FOV Angle", w.FovAngle.Val)
	fmt.Println("Eye Dist", w.EyeDist.Val)
}

type Segment struct {
	A, B mgl32.Vec3
	T    ObjectType
}

type Sphere struct {
	C mgl32.Vec3
	R float32
	T ObjectType
}

func MakeSegment(p1, p2 m3space.Point, t ObjectType) (Segment) {
	return Segment{
		mgl32.Vec3{float32(p1[0]), float32(p1[1]), float32(p1[2])},
		mgl32.Vec3{float32(p2[0]), float32(p2[1]), float32(p2[2])},
		t,
	}
}

func MakeSphere(c m3space.Point, t ObjectType) (Sphere) {
	return Sphere{
		mgl32.Vec3{float32(c[0]), float32(c[1]), float32(c[2])},
		SphereRadius.Val,
		t,
	}
}

type Triangle struct {
	vertices [pointsPerTriangle]mgl32.Vec3
	normal   mgl32.Vec3
	color    mgl32.Vec3
}

var AxeXColor = mgl32.Vec3{0.5, 0.2, 0.2}
var AxeYColor = mgl32.Vec3{0.2, 0.5, 0.2}
var AxeZColor = mgl32.Vec3{0.2, 0.2, 0.5}
var NodeAColor = mgl32.Vec3{1.0, 1.0, 0.0}
var NodeBColor = mgl32.Vec3{0.0, 1.0, 1.0}
var NodeCColor = mgl32.Vec3{1.0, 0.0, 1.0}
var Conn1Color = mgl32.Vec3{1.0, 0.0, 0.0}
var Conn9Color = mgl32.Vec3{0.0, 0.0, 1.0}

func squareDist(v mgl32.Vec3) float32 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

func MakeTriangle(vert [3]mgl32.Vec3, t ObjectType) Triangle {
	AB := vert[1].Sub(vert[0])
	AC := vert[2].Sub(vert[0])
	norm := AB.Cross(AC).Normalize()
	color := mgl32.Vec3{}
	switch t {
	case axeX:
		color = AxeXColor
	case axeY:
		color = AxeYColor
	case axeZ:
		color = AxeZColor
	case nodeA:
		color = NodeAColor
	case nodeB:
		color = NodeBColor
	case nodeC:
		color = NodeCColor
	case connection:
		squareConnLength := squareDist(AB)
		acL := squareDist(AC)
		if squareConnLength < acL {
			squareConnLength = acL
		}
		// Between 1 and 9
		color = Conn1Color.Mul(9 - squareConnLength).Add(Conn9Color.Mul(squareConnLength - 1)).Normalize()
	}
	return Triangle{vert, norm, color}
}

var XYZ = [3]mgl32.Vec3{{1.0, 0.0, 0.0}, {0.0, 1.0, 0.0}, {0.0, 0.0, 1.0}}

var Circle = []mgl32.Vec2{
	{1.0, 0.0},
	{1.0, 1.0},
	{0.0, 1.0},
	{-1.0, 1.0},
	{-1.0, 0.0},
	{-1.0, -1.0},
	{0.0, -1.0},
	{1.0, -1.0},
}

type GLObject interface {
	ExtractTriangles() []Triangle
}

func (s Sphere) ExtractTriangles() []Triangle {
	up := XYZ[2].Mul(s.R)
	south := s.C.Sub(up)
	north := s.C.Add(up)

	halfUp := XYZ[2].Mul(0.5)
	bottomCircle := make([]mgl32.Vec3, circleParts+1)
	equatorCircle := make([]mgl32.Vec3, 9)
	topCircle := make([]mgl32.Vec3, 9)
	for i, c := range Circle {
		equatorCircle[i] = XYZ[0].Mul(c[0]).Add(XYZ[1].Mul(c[1]))
		bottomCircle[i] = halfUp.Mul(-1.0).Add(equatorCircle[i])
		topCircle[i] = halfUp.Add(equatorCircle[i])
		equatorCircle[i] = equatorCircle[i].Mul(s.R / equatorCircle[i].Len())
		bottomCircle[i] = bottomCircle[i].Mul(s.R / bottomCircle[i].Len())
		topCircle[i] = topCircle[i].Mul(s.R / topCircle[i].Len())
		equatorCircle[i] = s.C.Add(equatorCircle[i])
		bottomCircle[i] = s.C.Add(bottomCircle[i])
		topCircle[i] = s.C.Add(topCircle[i])
	}
	equatorCircle[8] = equatorCircle[0]
	bottomCircle[8] = bottomCircle[0]
	topCircle[8] = topCircle[0]
	stride := 6
	result := make([]Triangle, trianglesPerSphere)
	for i := 0; i < 8; i++ {
		// South triangle
		result[stride*i] = MakeTriangle([3]mgl32.Vec3{
			south, bottomCircle[i+1], bottomCircle[i],
		}, s.T)
		// Bottom Triangles
		result[stride*i+1] = MakeTriangle([3]mgl32.Vec3{
			bottomCircle[i], equatorCircle[i+1], equatorCircle[i],
		}, s.T)
		result[stride*i+2] = MakeTriangle([3]mgl32.Vec3{
			equatorCircle[i+1], bottomCircle[i], bottomCircle[i+1],
		}, s.T)
		// Top Triangles
		result[stride*i+3] = MakeTriangle([3]mgl32.Vec3{
			equatorCircle[i], topCircle[i+1], topCircle[i],
		}, s.T)
		result[stride*i+4] = MakeTriangle([3]mgl32.Vec3{
			topCircle[i+1], equatorCircle[i], equatorCircle[i+1],
		}, s.T)
		// North triangle
		result[stride*i+5] = MakeTriangle([3]mgl32.Vec3{
			north, topCircle[i], topCircle[i+1],
		}, s.T)
	}
	return result
}

func (s Segment) ExtractTriangles() []Triangle {
	AB := s.B.Sub(s.A).Normalize()
	bestCross := mgl32.Vec3{0.0, 0.0, 0.0}
	for _, axe := range XYZ {
		cross := axe.Cross(AB)
		if cross.Len() > bestCross.Len() {
			bestCross = cross
		}
	}
	bestCross = bestCross.Normalize()
	cross2 := AB.Cross(bestCross).Normalize()
	// Let's draw a little cylinder around AB using bestCross and cross2 normal axes
	aPoints := make([]mgl32.Vec3, circleParts+1)
	bPoints := make([]mgl32.Vec3, circleParts+1)
	for i, c := range Circle {
		norm := bestCross.Mul(c[0]).Add(cross2.Mul(c[1])).Normalize().Mul(LineWidth.Val / 2.0)
		aPoints[i] = s.A.Add(norm)
		bPoints[i] = s.B.Add(norm)
	}
	// close the circle
	aPoints[circleParts] = aPoints[0]
	bPoints[circleParts] = bPoints[0]
	result := make([]Triangle, trianglesPerLine)
	for i := 0; i < circleParts; i++ {
		result[2*i] = MakeTriangle([3]mgl32.Vec3{
			aPoints[i], bPoints[i+1], bPoints[i],
		}, s.T)
		result[2*i+1] = MakeTriangle([3]mgl32.Vec3{
			bPoints[i+1], aPoints[i], aPoints[i+1],
		}, s.T)
	}
	return result
}

func (w *World) ScaleView(win *glfw.Window) int64 {
	w.Width, w.Height = win.GetSize()
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
		w.Angle += float32(elapsed / 2.0)
		w.Model = mgl32.HomogRotate3D(w.Angle, mgl32.Vec3{0, 1, 0})
	}
}

func (w *World) SetMatrices() {
	Eye := mgl32.Vec3{w.EyeDist.Val, w.EyeDist.Val, w.EyeDist.Val,}
	Far := Eye.Len() + w.TopCornerDist
	w.Projection = mgl32.Perspective(mgl32.DegToRad(w.FovAngle.Val), float32(w.Width)/float32(w.Height), 1.0, Far)
	w.Camera = mgl32.LookAtV(Eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
}

func (w *World) createAxesTriangles() {
	axeFiller := TriangleFiller{0, &w.AxesTriangles}
	for axe := int16(0); axe < axes; axe++ {
		p1 := m3space.Point{0, 0, 0}
		p2 := m3space.Point{0, 0, 0}
		p1[axe] = -w.Max
		p2[axe] = w.Max
		axeFiller.fill(MakeSegment(p1, m3space.Origin, ObjectType(axe)))
	}
}

func (w *World) createNodesAndConnectionsTriangles() {
	nodeFiller := TriangleFiller{0, &w.NodesTriangles}
	connectionFiller := TriangleFiller{0, &w.ConnectionsTriangles}
	nodeFiller.fill(MakeSphere(m3space.Origin, nodeA))
	for node := 1; node < nodes; node++ {
		nodeFiller.fill(MakeSphere(m3space.BasePoints[node-1], ObjectType(node+2)))
		connectionFiller.fill(MakeSegment(m3space.Origin, m3space.BasePoints[node-1], connection))
	}
}

func (w *World) FillAllVertices() {
	w.createAxesTriangles()
	w.createNodesAndConnectionsTriangles()
	w.NbTriangles = len(w.AxesTriangles) + len(w.NodesTriangles) + len(w.ConnectionsTriangles)
	if w.NbVertices != w.NbTriangles*3 {
		w.NbVertices = w.NbTriangles * 3
		fmt.Println("Creating OpenGL buffer for", w.NbTriangles, "triangles,", w.NbVertices, "vertices,", w.NbVertices*FloatPerVertices, "buffer size.")
		w.OpenGLBuffer = make([]float32, w.NbVertices*FloatPerVertices)
	}
	offset := w.fillOpenGlBuffer(w.AxesTriangles, 0)
	fmt.Println("After fill axes offset=", offset)
	offset = w.fillOpenGlBuffer(w.NodesTriangles, offset)
	fmt.Println("After fill nodes offset=", offset)
	offset = w.fillOpenGlBuffer(w.ConnectionsTriangles, offset)
	fmt.Println("After fill connections offset=", offset)
}

func (w *World) fillOpenGlBuffer(triangles []Triangle, offset int) int {
	for _, t := range triangles {
		for _, point := range t.vertices {
			for coord := 0; coord < coordinates; coord++ {
				w.OpenGLBuffer[offset] = point[coord]
				offset++
			}
			for coord := 0; coord < coordinates; coord++ {
				w.OpenGLBuffer[offset] = t.normal[coord]
				offset++
			}
			for coord := 0; coord < coordinates; coord++ {
				w.OpenGLBuffer[offset] = t.color[coord]
				offset++
			}
		}
	}
	return offset
}

type TriangleFiller struct {
	offset int
	array  *[]Triangle
}

func (t *TriangleFiller) fill(o GLObject) {
	triangles := o.ExtractTriangles()
	for _, tr := range triangles {
		(*t.array)[t.offset] = tr
		t.offset++
	}
}
