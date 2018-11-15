package m3space

import (
	"fmt"
)

type Point [3]int64

type GrowthContext struct {
	center             *Point
	permutationType    uint8 // 1,2,4, or 8
	permutationIndex   int   // Index in the permutations to choose from. For type 1 [0,7] for the other in the 12 list [0,11]
	permutationNegFlow bool  // true for backward flow in permutation
	permutationOffset  int   // Offset in perm modulo
}

var Origin = Point{0, 0, 0}
var XFirst = Point{THREE, 0, 0}
var YFirst = Point{0, THREE, 0}
var ZFirst = Point{0, 0, THREE}

func PosMod2(i int64) int64 {
	return i & 0x0000000000000001
}

func PosMod4(i int64) int64 {
	return i & 0x0000000000000003
}

func PosMod8(i int64) int64 {
	return i & 0x0000000000000007
}

func (ctx *GrowthContext) GetTrioIndex(divByThree int64) int {
	if ctx.permutationType == 1 {
		// Always same value
		return ctx.permutationIndex
	}
	if ctx.permutationType == 3 {
		// Center on Trio index ctx.permutationIndex and then use X, Y, Z where conn are 1
		mod2 := PosMod2(divByThree)
		if mod2 == 0 {
			return ctx.permutationIndex
		}
		mod3 := int(((divByThree-1)/2 + int64(ctx.permutationOffset)) % 3)
		if mod3 < 0 {
			mod3 += 3
		}
		if ctx.permutationIndex < 4 {
			return ValidNextTrio[3*ctx.permutationIndex+mod3][1]
		}
		count := 0
		for _, validTrio := range ValidNextTrio {
			if validTrio[1] == ctx.permutationIndex {
				if count == mod3 {
					return validTrio[0]
				}
				count++
			}
		}
		panic(fmt.Sprintf("did not find valid Trio for div by three value %d in context %v!", divByThree, ctx))
	}

	divByThreeWithOffset := int64(ctx.permutationOffset) + divByThree
	switch ctx.permutationType {
	case 2:
		permutMap := ValidNextTrio[ctx.permutationIndex]
		idx := int(PosMod2(divByThreeWithOffset))
		if ctx.permutationNegFlow {
			return permutMap[reverse2Map[idx]]
		} else {
			return permutMap[idx]
		}
	case 4:
		permutMap := AllMod4Permutations[ctx.permutationIndex]
		idx := int(PosMod4(divByThreeWithOffset))
		if ctx.permutationNegFlow {
			return permutMap[reverse4Map[idx]]
		} else {
			return permutMap[idx]
		}
	case 8:
		permutMap := AllMod8Permutations[ctx.permutationIndex]
		idx := int(PosMod8(divByThreeWithOffset))
		if ctx.permutationNegFlow {
			return permutMap[reverse8Map[idx]]
		} else {
			return permutMap[idx]
		}
	}
	panic(fmt.Sprintf("event permutation type %d in context %v is invalid!", ctx.permutationType, ctx))
}

func (ctx *GrowthContext) GetDivByThree(p Point) int64 {
	if !p.IsMainPoint() {
		panic(fmt.Sprintf("cannot ask for Trio index on non main Pos %v in context %v!", p, ctx))
	}
	return int64(Abs(p[0]-ctx.center[0])/3 + Abs(p[1]-ctx.center[1])/3 + Abs(p[2]-ctx.center[2])/3)
}

func (ctx *GrowthContext) GetTrio(p Point) Trio {
	return AllBaseTrio[ctx.GetTrioIndex(ctx.GetDivByThree(p))]
}

func (p Point) getNearMainPoint() Point {
	res := Point{}
	for i, c := range p {
		switch c % THREE {
		case 0:
			res[i] = c
		case 1:
			res[i] = c - 1
		case 2:
			res[i] = c + 1
		case -1:
			res[i] = c + 1
		case -2:
			res[i] = c - 1
		}
	}
	return res
}

// Give the 3 next points of a given node activated in the context of the current event.
// Return a clean new array not interacting with existing nodes, just the points extensions here based on the permutations.
// TODO (in the calling method): If the node already connected,
// TODO: only the connecting points that natches the normal event growth permutation cycle are returned.
func (currentPoint Point) getNextPoints(ctx *GrowthContext) [3]Point {
	result := [3]Point{}
	if currentPoint.IsMainPoint() {
		trio := ctx.GetTrio(currentPoint)
		for i, tr := range trio {
			result[i] = currentPoint.Add(tr)
		}
		return result
	}
	mainPoint := currentPoint.getNearMainPoint()
	result[0] = mainPoint
	cVec := currentPoint.Sub(mainPoint)
	nextPoints := mainPoint.getNextPointsFromMainAndVector(cVec, ctx)
	result[1] = nextPoints[0]
	result[2] = nextPoints[1]
	return result
}

func (mainPoint Point) getNextPointsFromMainAndVector(cVec Point, ctx *GrowthContext) [2]Point {
	offset := 0
	result := [2]Point{}

	nextMain := mainPoint
	switch cVec.X() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(XFirst)
	case -1:
		nextMain = mainPoint.Sub(XFirst)
	default:
		fmt.Printf("There should not be a connecting vector with x value %d\n", cVec.X())
		return result
	}
	if nextMain != mainPoint {
		// Find the base Pos on the other side ( the opposite 1 or -1 on X() )
		nextConnectingVectors := ctx.GetTrio(nextMain)
		for _, nbp := range nextConnectingVectors {
			if nbp.X() == -cVec.X() {
				result[offset] = nextMain.Add(nbp)
				offset++
			}
		}
	}

	nextMain = mainPoint
	switch cVec.Y() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(YFirst)
	case -1:
		nextMain = mainPoint.Sub(YFirst)
	default:
		fmt.Printf("There should not be a connecting vector with y value %d\n", cVec.Y())
	}
	if nextMain != mainPoint {
		// Find the base Pos on the other side ( the opposite 1 or -1 on Y() )
		nextConnectingVectors := ctx.GetTrio(nextMain)
		for _, nbp := range nextConnectingVectors {
			if nbp.Y() == -cVec.Y() {
				result[offset] = nextMain.Add(nbp)
				offset++
			}
		}
	}

	nextMain = mainPoint
	switch cVec.Z() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(ZFirst)
	case -1:
		nextMain = mainPoint.Sub(ZFirst)
	default:
		fmt.Printf("There should not be a connecting vector with z value %d\n", cVec.Z())
	}
	if nextMain != mainPoint {
		// Find the base Pos on the other side ( the opposite 1 or -1 on Z() )
		nextConnectingVectors := ctx.GetTrio(nextMain)
		for _, nbp := range nextConnectingVectors {
			if nbp.Z() == -cVec.Z() {
				result[offset] = nextMain.Add(nbp)
				offset++
			}
		}
	}
	return result
}

func Abs(i int64) int64 {
	if i < 0 {
		return -i
	}
	return i
}

func (p Point) X() int64 {
	return p[0]
}

func (p Point) Y() int64 {
	return p[1]
}

func (p Point) Z() int64 {
	return p[2]
}

func (p Point) Mul(m int64) Point {
	return Point{p[0] * m, p[1] * m, p[2] * m}
}

func (p1 Point) Add(p2 Point) Point {
	return Point{p1[0] + p2[0], p1[1] + p2[1], p1[2] + p2[2]}
}

func (p1 Point) Sub(p2 Point) Point {
	return Point{p1[0] - p2[0], p1[1] - p2[1], p1[2] - p2[2]}
}

func (p Point) Neg() Point {
	return Point{-p[0], -p[1], -p[2]}
}

// Positive PI/2 rotation on X
func (p Point) PlusX() Point {
	return Point{p[0], -p[2], p[1]}
}

// Negative PI/2 rotation on X
func (p Point) NegX() Point {
	return Point{p[0], p[2], -p[1]}
}

// Positive PI/2 rotation on Y
func (p Point) PlusY() Point {
	return Point{p[2], p[1], -p[0]}
}

// Negative PI/2 rotation on Y
func (p Point) NegY() Point {
	return Point{-p[2], p[1], p[0]}
}

// Positive PI/2 rotation on Z
func (p Point) PlusZ() Point {
	return Point{-p[1], p[0], p[2]}
}

// Negative PI/2 rotation on X
func (p Point) NegZ() Point {
	return Point{p[1], -p[0], p[2]}
}

func DS(p1, p2 *Point) int64 {
	x := p2.X() - p1.X()
	y := p2.Y() - p1.Y()
	z := p2.Z() - p1.Z()
	return x*x + y*y + z*z
}

func (p Point) DistanceSquared() int64 {
	return p[0]*p[0] + p[1]*p[1] + p[2]*p[2]
}

func (p Point) IsMainPoint() bool {
	allDivByThree := true
	for _, c := range p {
		if c%THREE != 0 {
			allDivByThree = false
		}
	}
	return allDivByThree
}

func (p Point) IsBaseConnectingVector() bool {
	if p.IsOnlyOneAndZero() {
		return p.DistanceSquared() == 2
	}
	return false
}

func (p Point) IsConnectionVector() bool {
	if p.IsOnlyTwoOneAndZero() {
		sizeSquared := p.DistanceSquared()
		return sizeSquared == 1 || sizeSquared == 2 || sizeSquared == 3 || sizeSquared == 5
	}
	return false
}

func (p Point) IsOnlyOneAndZero() bool {
	for _, c := range p {
		if c != 0 && c != 1 && c != -1 {
			return false
		}
	}
	return true
}

func (p Point) IsOnlyTwoOneAndZero() bool {
	for _, c := range p {
		if c != 0 && c != 1 && c != -1 && c != 2 && c != -2 {
			return false
		}
	}
	return true
}

func (p Point) IsBorder(max int64) bool {
	for _, c := range p {
		if c > 0 && c >= max-1 {
			return true
		}
		if c < 0 && c <= -max+1 {
			return true
		}
	}
	return false
}

func (p Point) IsOutBorder(max int64) bool {
	for _, c := range p {
		if c > 0 && c > max {
			return true
		}
		if c < 0 && c < -max {
			return true
		}
	}
	return false
}
