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
	circlePartsLine    = 8
	trianglesPerLine   = circlePartsLine * 2
	circlePartsSphere  = 32
	nbMiddleCircles    = (circlePartsSphere - 2) / 2
	trianglesPerSphere = nbMiddleCircles * circlePartsSphere * 2
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
	verifyData()
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
		mgl32.Vec3{0.0, 1.0, 1.0}.Normalize(),
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
		CircleForSphere[i] = mgl32.Vec2{float32(math.Cos(deltaAngle * float64(i))), float32(math.Sin(deltaAngle * float64(i)))}
	}
}

func (w World) DisplaySettings() {
	fmt.Println("========= World Settings =========")
	fmt.Println("Width", w.Width, "Height", w.Height)
	fmt.Println("Line Width [B,T]", LineWidth.Val)
	fmt.Println("Sphere Radius [P,L]", SphereRadius.Val)
	fmt.Println("FOV Angle [Z,X]", w.FovAngle.Val)
	fmt.Println("Eye Dist [Q,W]", w.EyeDist.Val)
}

func (w *World) createAxesTriangles() {
	axeFiller := TriangleFiller{0, &w.AxesTriangles}
	for axe := int16(0); axe < axes; axe++ {
		p1 := m3space.Point{0, 0, 0}
		p2 := m3space.Point{0, 0, 0}
		p1[axe] = -w.Max
		p2[axe] = w.Max
		axeFiller.fill(MakeSegment(p1, m3space.Origin, ObjectType(axe)))
		axeFiller.fill(MakeSegment(m3space.Origin, p2, ObjectType(axe)))
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
	ExtractTriangles() []Triangle
}

func (s Sphere) ExtractTriangles() []Triangle {
	up := ZH.Mul(s.R)
	south := s.C.Sub(up)
	north := s.C.Add(up)

	middleCircles := make([][]mgl32.Vec3, nbMiddleCircles)
	middleCirclesNorm := make([][]mgl32.Vec3, nbMiddleCircles)
	middleCirclesZPart := make([]float32, nbMiddleCircles)
	for z := 0; z < nbMiddleCircles; z++ {
		middleCirclesZPart[z] = -CircleForSphere[z+1].Normalize().X()
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
	w.Width, w.Height = win.GetSize()
	gl.Viewport(0, 0, int32(w.Width*2), int32(w.Height*2))
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
	offset = w.fillOpenGlBuffer(w.NodesTriangles, offset)
	offset = w.fillOpenGlBuffer(w.ConnectionsTriangles, offset)
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
