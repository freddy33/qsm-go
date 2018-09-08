package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/v4.1-core/gl"
	"math"
)

// Environment const
const (
	RetinaDisplay = true
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

// QSM Objects const
const (
	nodes       = 3
	connections = 2
	axes        = 3
)

type ObjectType int16
type ObjectKey int

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
	Max           int64
	TopCornerDist float32

	NbVertices   int
	OpenGLBuffer []float32
	Objects      map[ObjectKey]WorldObject

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
	Angle          float32
	Rotate         bool
}

type WorldObject struct {
	k            ObjectKey
	OpenGLOffset int32
	NbVertices   int32
}

func MakeWorld(Max int64) World {
	verifyData()
	TopCornerDist := float32(math.Sqrt(float64(3.0*Max*Max))) + 1.1
	w := World{
		Max,
		TopCornerDist,
		0,
		make([]float32, 0),
		make(map[ObjectKey]WorldObject),
		800, 600,
		SizeVar{float32(Max), TopCornerDist * 4.0, TopCornerDist},
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
	return w
}

var LineWidth = SizeVar{0.001, 0.5, 0.04}
var SphereRadius = SizeVar{0.05, 0.5, 0.1}
var XH = mgl32.Vec3{1.0, 0.0, 0.0}
var YH = mgl32.Vec3{0.0, 1.0, 0.0}
var ZH = mgl32.Vec3{0.0, 0.0, 1.0}
var XYZ = [3]mgl32.Vec3{XH, YH, ZH}
var CircleForLine = make([]mgl32.Vec2, circlePartsLine)
var CircleForSphere = make([]mgl32.Vec2, circlePartsSphere)

func verifyData() {
	// Verify we capture the equator
	if nbMiddleCircles%2 == 0 {
		panic(fmt.Errorf("something fishy with circle parts %d since %d should be odd", circlePartsSphere, nbMiddleCircles))
	}
	deltaAngle := 2.0 * math.Pi / circlePartsLine
	for i := 0; i < circlePartsLine; i++ {
		CircleForLine[i] = mgl32.Vec2{float32(math.Cos(deltaAngle * float64(i))), float32(math.Sin(deltaAngle * float64(i)))}
	}
	deltaAngle = 2.0 * math.Pi / circlePartsSphere
	for i := 0; i < circlePartsSphere; i++ {
		angle := deltaAngle * float64(i)
		CircleForSphere[i] = mgl32.Vec2{float32(math.Cos(angle)), float32(math.Sin(angle))}
	}
}

func (w World) DisplaySettings() {
	fmt.Println("========= World Settings =========")
	fmt.Println("Nb Objects", len(w.Objects))
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
	triangleFiller := TriangleFiller{make(map[ObjectKey]WorldObject), 0, 0,&(w.OpenGLBuffer)}
	p := m3space.Point{w.Max, 0, 0}
	for axe := int16(0); axe < axes; axe++ {
		triangleFiller.fill(MakeSegment(m3space.Origin, p, ObjectType(axe)))
	}
	for node := 0; node < nodes; node++ {
		triangleFiller.fill(MakeSphere(ObjectType(int(nodeA) + node)))
	}
	triangleFiller.fill(MakeSegment(m3space.Origin, m3space.BasePoints[0], connection))
	triangleFiller.fill(MakeSegment(m3space.BasePoints[0], m3space.BasePoints[1].Add(m3space.Point{0, 3, 0}), connection))

	w.Objects = triangleFiller.objMap

	return nbTriangles
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

var Origin = mgl32.Vec3{0.0, 0.0, 0.0}

func MakeSegment(p1, p2 m3space.Point, t ObjectType) (Segment) {
	length := float32(math.Sqrt(float64(m3space.DS(p1, p2))))
	return Segment{
		Origin,
		mgl32.Vec3{length, 0.0, 0.0},
		t,
	}
}

func MakeSphere(t ObjectType) (Sphere) {
	return Sphere{
		Origin,
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

func MakeTriangleWithNorm(vert [3]mgl32.Vec3, norm mgl32.Vec3, t ObjectType) Triangle {
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
		AB := vert[1].Sub(vert[0])
		AC := vert[2].Sub(vert[0])
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

func MakeTriangle(vert [3]mgl32.Vec3, t ObjectType) Triangle {
	AB := vert[1].Sub(vert[0])
	AC := vert[2].Sub(vert[0])
	norm := AB.Cross(AC).Normalize()
	return MakeTriangleWithNorm(vert, norm, t)
}

type GLObject interface {
	Size() int
	Key() ObjectKey
	NumberOfVertices() int
	ExtractTriangles() []Triangle
}

func (s Sphere) Size() int {
	return int(s.R * 1000)
}

func (s Sphere) Key() ObjectKey {
	return ObjectKey(int(s.T) + s.Size()*100)
}

func (s Sphere) NumberOfVertices() int {
	return trianglesPerSphere * pointsPerTriangle
}

func (s Sphere) ExtractTriangles() []Triangle {
	up := ZH.Mul(s.R)
	south := s.C.Sub(up)
	north := s.C.Add(up)

	middleCircles := make([][]mgl32.Vec3, nbMiddleCircles)
	middleCirclesNorm := make([][]mgl32.Vec3, nbMiddleCircles)
	middleCirclesZPart := make([]float32, nbMiddleCircles)
	deltaAngle := 2.0 * math.Pi / circlePartsSphere
	for z := 0; z < nbMiddleCircles; z++ {
		angle := deltaAngle * float64(z+1)
		if z == 0 {
			angle = deltaAngle / 4
			//angle -= 3*deltaAngle/2
		}
		if z == nbMiddleCircles-1 {
			angle += 3 * deltaAngle / 4
		}
		middleCirclesZPart[z] = -float32(math.Cos(angle))
		middleCircles[z] = make([]mgl32.Vec3, circlePartsSphere+1)
		middleCirclesNorm[z] = make([]mgl32.Vec3, circlePartsSphere)
	}
	for i, c := range CircleForSphere {
		for zIdx, zH := range middleCirclesZPart {
			middleCirclesNorm[zIdx][i] = mgl32.Vec3{c[0], c[1], zH}.Normalize()
			middleCircles[zIdx][i] = s.C.Add(middleCirclesNorm[zIdx][i].Mul(s.R))
		}
	}
	for zIdx := range middleCircles {
		middleCircles[zIdx][circlePartsSphere] = middleCircles[zIdx][0]
	}

	offset := 0
	result := make([]Triangle, trianglesPerSphere)
	for i := 0; i < circlePartsSphere; i++ {
		// South triangle
		result[offset] = MakeTriangleWithNorm([3]mgl32.Vec3{
			south, middleCircles[0][i+1], middleCircles[0][i],
		}, middleCirclesNorm[0][i], s.T)
		offset++
		for zIdx := 0; zIdx < nbMiddleCircles-1; zIdx++ {
			result[offset] = MakeTriangleWithNorm([3]mgl32.Vec3{
				middleCircles[zIdx][i], middleCircles[zIdx][i+1], middleCircles[zIdx+1][i+1],
			}, middleCirclesNorm[zIdx][i], s.T)
			offset++
			result[offset] = MakeTriangleWithNorm([3]mgl32.Vec3{
				middleCircles[zIdx+1][i+1], middleCircles[zIdx+1][i], middleCircles[zIdx][i],
			}, middleCirclesNorm[zIdx][i], s.T)
			offset++
		}
		// North triangle
		result[offset] = MakeTriangleWithNorm([3]mgl32.Vec3{
			north, middleCircles[nbMiddleCircles-1][i], middleCircles[nbMiddleCircles-1][i+1],
		}, middleCirclesNorm[nbMiddleCircles-1][i], s.T)
		offset++
	}
	return result
}

func (s Segment) Size() int {
	return int(s.A.Sub(s.B).Len() * 1000)
}

func (s Segment) Key() ObjectKey {
	return ObjectKey(int(s.T) + s.Size()*100)
}

func (s Segment) NumberOfVertices() int {
	return trianglesPerLine * pointsPerTriangle
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
	aPoints := make([]mgl32.Vec3, circlePartsLine+1)
	bPoints := make([]mgl32.Vec3, circlePartsLine+1)
	for i, c := range CircleForLine {
		norm := bestCross.Mul(c[0]).Add(cross2.Mul(c[1])).Normalize().Mul(LineWidth.Val / 2.0)
		aPoints[i] = s.A.Add(norm)
		bPoints[i] = s.B.Add(norm)
	}
	// close the circle
	aPoints[circlePartsLine] = aPoints[0]
	bPoints[circlePartsLine] = bPoints[0]
	result := make([]Triangle, trianglesPerLine)
	for i := 0; i < circlePartsLine; i++ {
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
		w.Angle += float32(elapsed / 2.0)
		w.Model = mgl32.HomogRotate3D(w.Angle, mgl32.Vec3{0, 0, 1})
	}
}

func (w *World) SetMatrices() {
	Eye := mgl32.Vec3{w.EyeDist.Val, w.EyeDist.Val, w.EyeDist.Val,}
	Far := Eye.Len() + w.TopCornerDist
	w.Projection = mgl32.Perspective(mgl32.DegToRad(w.FovAngle.Val), float32(w.Width)/float32(w.Height), 1.0, Far)
	w.Camera = mgl32.LookAtV(Eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, 1})
}

type TriangleFiller struct {
	objMap           map[ObjectKey]WorldObject
	verticesOffset int32
	bufferOffset int
	buffer       *[]float32
}

func (t *TriangleFiller) fill(o GLObject) {
	wo := WorldObject{
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
				(*t.buffer)[t.bufferOffset] = point[coord]
				t.bufferOffset++
			}
			for coord := 0; coord < coordinates; coord++ {
				(*t.buffer)[t.bufferOffset] = tr.normal[coord]
				t.bufferOffset++
			}
			for coord := 0; coord < coordinates; coord++ {
				(*t.buffer)[t.bufferOffset] = tr.color[coord]
				t.bufferOffset++
			}
		}
	}
}
