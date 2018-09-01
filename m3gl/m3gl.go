package m3gl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/freddy33/qsm-go/m3space"
	"fmt"
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
var Xvec = mgl32.Vec3{1.0,0.0,0.0}
var Yvec = mgl32.Vec3{0.0,1.0,0.0}

func (s Segment) ExtractTriangles() ([4]Triangle, error) {
	AB := s.B.Sub(s.A)
	cross1 := Xvec.Cross(AB)
	if cross1.Len() < 0.001 {
		cross1 = Yvec.Cross(AB)
	}
	if cross1.Len() < 0.001 {
		return [4]Triangle {}, fmt.Errorf("did not find cross vector big enough for %v", AB)
	}
	cross1 = cross1.Normalize()
	cross2 := cross1.Cross(AB).Normalize().Mul(lineWidth/2.0)
	cross1 = cross1.Mul(lineWidth/2.0)

	A11 := s.A.Sub(cross1)
	A12 := s.A.Add(cross1)
	B11 := s.B.Sub(cross1)
	B12 := s.B.Add(cross1)
	A21 := s.A.Sub(cross2)
	A22 := s.A.Add(cross2)
	B21 := s.B.Sub(cross2)
	B22 := s.B.Add(cross2)
	return [4]Triangle {
		{
			[3]mgl32.Vec3 {
				A11,
				B11,
				B12,
			},
		},
		{
			[3]mgl32.Vec3 {
				B12,
				A12,
				A11,
			},
		},
		{
			[3]mgl32.Vec3 {
				A21,
				B21,
				B22,
			},
		},
		{
			[3]mgl32.Vec3 {
				B22,
				A22,
				A21,
			},
		},
	}, nil
}
