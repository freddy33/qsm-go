package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
)

type GrowthContext struct {
	m3point.TrioIndexContext
	// Where this growth context starts from
	center m3point.Point
	// Offset in permutation to start with. Basically index in perm pos at center
	offset int
}

type PathContext struct {
	ctx           *GrowthContext
	trioSequences []PathContextElement
}

type PathContextElement struct {
	srcTrio   *m3point.TrioDetails
	nextTrios []*m3point.TrioDetails
}

func CreateGrowthContext(center m3point.Point, permType m3point.ContextType, index int, offset int) *GrowthContext {
	return &GrowthContext{*m3point.GetTrioIndexContext(permType, index), center, offset}
}

var maxOffsetPerType = map[m3point.ContextType]int{
	m3point.ContextType(1): 1,
	m3point.ContextType(3): 4,
	m3point.ContextType(2): 2,
	m3point.ContextType(4): 4,
	m3point.ContextType(8): 8,
}

func CreateFromRoot(trioIndexCtx *m3point.TrioIndexContext, center m3point.Point, offset int) *GrowthContext {
	if offset < 0 || offset >= maxOffsetPerType[trioIndexCtx.GetType()] {
		Log.Error("Offset value %d invalid for context type", offset, trioIndexCtx.GetType())
		return nil
	}
	return CreateGrowthContext(center, trioIndexCtx.GetType(), trioIndexCtx.GetIndex(), offset)
}

func (ctx *GrowthContext) SetIndexOffset(idx, offset int) {
	ctx.SetIndex(idx)
	ctx.offset = offset
}

func (ctx *GrowthContext) SetCenter(c m3point.Point) {
	ctx.center = c
}

func (ctx *GrowthContext) GetCenter() m3point.Point {
	return ctx.center
}

func (ctx *GrowthContext) GetFileName() string {
	return fmt.Sprintf("C_%03d_%03d_%03d_G_%d_%d",
		ctx.center[0], ctx.center[1], ctx.center[2],
		ctx.GetType(), ctx.GetIndex())
}

func (ctx *GrowthContext) String() string {
	return fmt.Sprintf("GrowthType%d-Idx%d-Offset%d", ctx.GetType(), ctx.GetIndex(), ctx.offset)
}

func (ctx *GrowthContext) GetTrioIndex(divByThree uint64) int {
	return ctx.GetBaseTrioIndex(divByThree, ctx.offset)
}

func (ctx *GrowthContext) GetDivByThree(p m3point.Point) uint64 {
	if !p.IsMainPoint() {
		panic(fmt.Sprintf("cannot ask for Trio index on non main Pos %v in context %v!", p, ctx))
	}
	return uint64(m3point.Abs64(p[0]-ctx.center[0])/3 + m3point.Abs64(p[1]-ctx.center[1])/3 + m3point.Abs64(p[2]-ctx.center[2])/3)
}

func (ctx *GrowthContext) GetTrio(p m3point.Point) m3point.Trio {
	return m3point.AllBaseTrio[ctx.GetTrioIndex(ctx.GetDivByThree(p))]
}

// Give the 3 next points of a given node activated in the context of the current event.
// Return a clean new array not interacting with existing nodes, just the points extensions here based on the permutations.
// TODO (in the calling method): If the node already connected,
// TODO: only the connecting points that natches the normal event growth permutation cycle are returned.
func (ctx *GrowthContext) GetNextPoints(p m3point.Point) [3]m3point.Point {
	result := [3]m3point.Point{}
	if p.IsMainPoint() {
		trio := ctx.GetTrio(p)
		for i, tr := range trio {
			result[i] = p.Add(tr)
		}
		return result
	}
	mainPoint := p.GetNearMainPoint()
	result[0] = mainPoint
	cVec := p.Sub(mainPoint)
	nextPoints := getNextPointsFromMainAndVector(mainPoint, cVec, ctx)
	result[1] = nextPoints[0]
	result[2] = nextPoints[1]
	return result
}

/***************************************************************/
// Point Functions for only main points (all coord dividable by 3)
// TODO: Make MainPoint a type
/***************************************************************/

func getNextPointsFromMainAndVector(mainPoint m3point.Point, cVec m3point.Point, ctx *GrowthContext) [2]m3point.Point {
	if !cVec.IsBaseConnectingVector() {
		Log.Fatalf("cannot do getNextPointsFromMainAndVector if %v not main base vector", cVec)
	}
	offset := 0
	result := [2]m3point.Point{}

	nextMain := mainPoint
	switch cVec.X() {
	case 0:
		// Nothing out
	case 1:
		nextMain = mainPoint.Add(m3point.XFirst)
	case -1:
		nextMain = mainPoint.Sub(m3point.XFirst)
	default:
		Log.Errorf("There should not be a connecting vector with x value %d\n", cVec.X())
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
		nextMain = mainPoint.Add(m3point.YFirst)
	case -1:
		nextMain = mainPoint.Sub(m3point.YFirst)
	default:
		Log.Errorf("There should not be a connecting vector with y value %d\n", cVec.Y())
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
		nextMain = mainPoint.Add(m3point.ZFirst)
	case -1:
		nextMain = mainPoint.Sub(m3point.ZFirst)
	default:
		Log.Errorf("There should not be a connecting vector with z value %d\n", cVec.Z())
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
