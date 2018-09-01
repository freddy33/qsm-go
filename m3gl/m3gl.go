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
	return Segment{
		mgl32.Vec3{float32(p1[0]), float32(p1[1]), float32(p1[2])},
		mgl32.Vec3{float32(p2[0]), float32(p2[1]), float32(p2[2])},
	}
}

type Triangle struct {
	Points [3]mgl32.Vec3
}

var lineWidth = float32(0.02)

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
		norm := bestCross.Mul(c[0]).Add(cross2.Mul(c[1])).Normalize().Mul(lineWidth / 2.0)
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
