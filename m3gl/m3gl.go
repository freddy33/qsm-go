package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
	"fmt"
	"math"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Segment struct {
	A, B mgl32.Vec3
}

type Sphere struct {
	C mgl32.Vec3
	R float32
}

func MakeSegment(p1, p2 m3space.Point) (Segment) {
	return Segment{
		mgl32.Vec3{float32(p1[0]), float32(p1[1]), float32(p1[2])},
		mgl32.Vec3{float32(p2[0]), float32(p2[1]), float32(p2[2])},
	}
}

type Triangle struct {
	Points [3]mgl32.Vec3
}

var LineWidth = float32(0.06)

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

func (s Segment) ExtractTriangles() ([]Triangle, error) {
	AB := s.B.Sub(s.A)
	bestCross := mgl32.Vec3{0.0, 0.0, 0.0}
	for _, axe := range XYZ {
		cross := axe.Cross(AB)
		if cross.Len() > bestCross.Len() {
			bestCross = cross
		}
	}
	if bestCross.Len() < 0.001 {
		return []Triangle{}, fmt.Errorf("did not find cross vector big enough for %v", AB)
	}
	bestCross = bestCross.Normalize()
	cross2 := bestCross.Cross(AB).Normalize()
	// Let's draw a little cylinder around AB using bestCross and cross2 normal axes
	aPoints := make([]mgl32.Vec3, 9)
	bPoints := make([]mgl32.Vec3, 9)
	for i, c := range Circle {
		norm := bestCross.Mul(c[0]).Add(cross2.Mul(c[1])).Normalize().Mul(LineWidth / 2.0)
		aPoints[i] = s.A.Add(norm)
		bPoints[i] = s.B.Add(norm)
	}
	// close the circle
	aPoints[8] = aPoints[0]
	bPoints[8] = bPoints[0]
	result := make([]Triangle, 16)
	for i := 0; i < 8; i++ {
		result[2*i] = Triangle{[3]mgl32.Vec3{
			aPoints[i], bPoints[i], bPoints[i+1],
		}}
		result[2*i+1] = Triangle{[3]mgl32.Vec3{
			bPoints[i+1], aPoints[i+1], aPoints[i],
		}}
	}
	return result, nil
}

const (
	axes              = 3
	trianglePerLine   = 16
	pointsPerTriangle = 3
	coordinates       = 3
)

type World struct {
	Max                       int64
	NbVertices                int
	AxesVertices              []float32
	AxesType                  []int16
	Width, Height             int
	Eye                       mgl32.Vec3
	FovAngle                  float32
	Far                       float32
	Projection, Camera, Model mgl32.Mat4
	previousTime              float64
	previousArea              int64
	Angle                     float32
	Rotate                    bool
}

func MakeWorld(Max int64) World {
	eyeFromOrig := float32(math.Sqrt(float64(3.0*Max*Max))) + 1.1
	far := eyeFromOrig * 2.2
	eye := mgl32.Vec3{eyeFromOrig, eyeFromOrig, eyeFromOrig}
	nbVertices := 2 * axes * trianglePerLine * pointsPerTriangle
	w := World{
		Max,
		nbVertices,
		make([]float32, nbVertices*coordinates),
		make([]int16, nbVertices),
		800, 600,
		eye,
		45.0,
		far,
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
	w.Projection = mgl32.Perspective(mgl32.DegToRad(w.FovAngle), float32(w.Width)/float32(w.Height), 1.0, w.Far)
	w.Camera = mgl32.LookAtV(w.Eye, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
}

func (w *World) CreateAxes() {
	s := SegmentsVertices{0, &(w.AxesVertices), 0, &(w.AxesType)}
	o := m3space.Point{0, 0, 0}
	for axe := int16(0); axe < axes; axe++ {
		p1 := m3space.Point{0, 0, 0}
		p2 := m3space.Point{0, 0, 0}
		p1[axe] = -w.Max
		p2[axe] = w.Max
		color := mgl32.Vec3{1.0, 1.0, 1.0}
		color[axe] = 0.5
		s.fillVertices(MakeSegment(p1, o), axe)
		s.fillVertices(MakeSegment(o, p2), axe)
	}
}

type SegmentsVertices struct {
	pointOffset int
	pointArray  *[]float32
	typeOffset  int
	typeArray   *[]int16
}

func (s *SegmentsVertices) fillVertices(segment Segment, segmentType int16) {
	triangles, err := segment.ExtractTriangles()
	if err != nil {
		panic(err)
	}
	if len(triangles) != trianglePerLine {
		panic(fmt.Sprint("Number of triangles per lines inconsistent", len(triangles), trianglePerLine))
	}
	for _, triangle := range triangles {
		for _, point := range triangle.Points {
			(*s.typeArray)[s.typeOffset] = segmentType
			s.typeOffset++
			for coord := 0; coord < coordinates; coord++ {
				(*s.pointArray)[s.pointOffset] = point[coord]
				s.pointOffset++
			}
		}
	}
}
