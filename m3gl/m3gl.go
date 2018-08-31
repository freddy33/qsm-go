package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
)

type Segment struct {
	A, B mgl32.Vec3
}

func MakeSegment(p1, p2 m3space.Point) (Segment) {
	return Segment {
		mgl32.Vec3 {float32(p1[0]), float32(p1[1]), float32(p1[2])},
		mgl32.Vec3 {float32(p2[0]), float32(p2[1]), float32(p2[2])},
	}
}

type Triangle struct {
	Points [3]mgl32.Vec3
}

var lineWidth = float32(0.02)

func (s Segment) ExtractTriangles(eye mgl32.Vec3) ([2]Triangle) {
	AEye := eye.Sub(s.A)
	AB := s.B.Sub(s.A)
	cross := AEye.Cross(AB)
	var norm mgl32.Vec3
	if cross.Len() > 0.01 {
		norm = cross.Normalize().Mul(lineWidth/2.0)
	} else {
		AEye = s.B.Add(mgl32.Vec3{1.0,1.0,1.0}).Sub(s.A)
		cross = AEye.Cross(AB)
		norm = cross.Normalize().Mul(lineWidth/2.0)
	}
	A1 := s.A.Sub(norm)
	A2 := s.A.Add(norm)
	B1 := s.B.Sub(norm)
	B2 := s.B.Add(norm)
	return [2]Triangle {
		{
			[3]mgl32.Vec3 {
				A1,
				B1,
				B2,
			},
		},
		{
			[3]mgl32.Vec3 {
				B2,
				A2,
				A1,
			},
		},
	}
}
