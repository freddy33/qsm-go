package m3gl

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

type Segment struct {
	A, B mgl64.Vec3
	T    ObjectType
}

type Sphere struct {
	C mgl64.Vec3
	R float64
	T ObjectType
}

var Origin = mgl64.Vec3{0.0, 0.0, 0.0}

func MakeSegment(p1, p2 m3point.Point, t ObjectType) Segment {
	return Segment{
		Origin,
		mgl64.Vec3{float64(p2.X() - p1.X()), float64(p2.Y() - p1.Y()), float64(p2.Z() - p1.Z()),},
		t,
	}
}

func MakeSphere(t ObjectType) (Sphere) {
	if t == NodeEmpty {
		return Sphere{
			Origin,
			SphereRadius.Val / 2.0,
			t,
		}
	}
	return Sphere{
		Origin,
		SphereRadius.Val,
		t,
	}
}

type GLObject interface {
	Key() ObjectType
	NumberOfVertices() int
	ExtractTriangles() []Triangle
}

func (s Sphere) Key() ObjectType {
	return s.T
}

func (s Sphere) NumberOfVertices() int {
	return trianglesPerSphere * pointsPerTriangle
}

func (s Sphere) ExtractTriangles() []Triangle {
	up := ZH.Mul(s.R)
	south := s.C.Sub(up)
	north := s.C.Add(up)

	middleCircles := make([][]mgl64.Vec3, nbMiddleCircles)
	middleCirclesNorm := make([][]mgl64.Vec3, nbMiddleCircles)
	middleCirclesZPart := make([]mgl64.Vec2, nbMiddleCircles)
	deltaAngle := 2.0 * math.Pi / circlePartsSphere
	angle := deltaAngle
	for z := 0; z < nbMiddleCircles; z++ {
		middleCirclesZPart[z] = mgl64.Vec2{math.Sin(angle), -math.Cos(angle),}
		middleCircles[z] = make([]mgl64.Vec3, circlePartsSphere+1)
		middleCirclesNorm[z] = make([]mgl64.Vec3, circlePartsSphere)
		angle += deltaAngle
	}
	for i, c := range CircleForSphere {
		for zIdx, zH := range middleCirclesZPart {
			middleCirclesNorm[zIdx][i] = mgl64.Vec3{c[0] * zH[0], c[1] * zH[0], zH[1]}.Normalize()
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
		result[offset] = MakeTriangleWithNorm([3]mgl64.Vec3{
			south, middleCircles[0][i+1], middleCircles[0][i],
		}, middleCirclesNorm[0][i])
		offset++
		for zIdx := 0; zIdx < nbMiddleCircles-1; zIdx++ {
			result[offset] = MakeTriangleWithNorm([3]mgl64.Vec3{
				middleCircles[zIdx][i], middleCircles[zIdx][i+1], middleCircles[zIdx+1][i+1],
			}, middleCirclesNorm[zIdx][i])
			offset++
			result[offset] = MakeTriangleWithNorm([3]mgl64.Vec3{
				middleCircles[zIdx+1][i+1], middleCircles[zIdx+1][i], middleCircles[zIdx][i],
			}, middleCirclesNorm[zIdx][i])
			offset++
		}
		// North triangle
		result[offset] = MakeTriangleWithNorm([3]mgl64.Vec3{
			north, middleCircles[nbMiddleCircles-1][i], middleCircles[nbMiddleCircles-1][i+1],
		}, middleCirclesNorm[nbMiddleCircles-1][i])
		offset++
	}
	return result
}

func (s Segment) Key() ObjectType {
	return s.T
}

func (s Segment) NumberOfVertices() int {
	return trianglesPerLine * pointsPerTriangle
}

func (s Segment) ExtractTriangles() []Triangle {
	AB := s.B.Sub(s.A).Normalize()
	bestCross := mgl64.Vec3{0.0, 0.0, 0.0}
	for _, axe := range XYZ {
		cross := axe.Cross(AB)
		if cross.Len() > bestCross.Len() {
			bestCross = cross
		}
	}
	bestCross = bestCross.Normalize()
	cross2 := AB.Cross(bestCross).Normalize()
	// Let's draw a little cylinder around AB using bestCross and cross2 normal axes
	aPoints := make([]mgl64.Vec3, circlePartsLine+1)
	bPoints := make([]mgl64.Vec3, circlePartsLine+1)
	var lw float64
	if int(s.T) <= int(AxeZ) {
		lw = LineWidth.Val
	} else {
		lw = LineWidth.Val / 2.0
	}
	for i, c := range CircleForLine {
		norm := bestCross.Mul(c[0]).Add(cross2.Mul(c[1])).Normalize().Mul(lw)
		aPoints[i] = s.A.Add(norm)
		bPoints[i] = s.B.Add(norm)
	}
	// close the circle
	aPoints[circlePartsLine] = aPoints[0]
	bPoints[circlePartsLine] = bPoints[0]
	result := make([]Triangle, trianglesPerLine)
	for i := 0; i < circlePartsLine; i++ {
		result[2*i] = MakeTriangle([3]mgl64.Vec3{
			aPoints[i], bPoints[i+1], bPoints[i],
		})
		result[2*i+1] = MakeTriangle([3]mgl64.Vec3{
			bPoints[i+1], aPoints[i], aPoints[i+1],
		})
	}
	return result
}
