package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
	"math"
)

type Segment struct {
	A, B mgl32.Vec3
	T    m3space.ObjectType
	S    int
}

type Sphere struct {
	C mgl32.Vec3
	R float32
	T m3space.ObjectType
}

var Origin = mgl32.Vec3{0.0, 0.0, 0.0}

func MakeSegment(p1, p2 m3space.Point, t m3space.ObjectType) (Segment) {
	ds := int(m3space.DS(p1, p2))
	length := float32(math.Sqrt(float64(ds)))
	return Segment{
		Origin,
		mgl32.Vec3{length, 0.0, 0.0},
		t,
		ds,
	}
}

func MakeSphere(t m3space.ObjectType) (Sphere) {
	return Sphere{
		Origin,
		SphereRadius.Val,
		t,
	}
}

type GLObject interface {
	Size() int
	Key() m3space.ObjectKey
	NumberOfVertices() int
	ExtractTriangles() []Triangle
}

func (s Sphere) Size() int {
	return 1
}

func (s Sphere) Key() m3space.ObjectKey {
	return m3space.ObjectKey(int(s.T) + s.Size()*100)
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
	return s.S
}

func (s Segment) Key() m3space.ObjectKey {
	return m3space.ObjectKey(int(s.T) + s.Size()*100)
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
