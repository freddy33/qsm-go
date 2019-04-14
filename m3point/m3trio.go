package m3point

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	"sort"
	"strings"
)

// Trio of connection vectors from any point using connections only
type Trio [3]Point

// Keeping track of how base trio connects
type TrioLink struct {
	// The trio index of the source
	a int
	// The 2 possible trio index of the destination
	b, c int
}

// Defining a list type to manage uniqueness and ordering
type TrioLinkList []*TrioLink

// A bigger struct than Trio to keep more info on how points grow from a trio index
type TrioDetails struct {
	id    int
	conns [3]*ConnectionDetails
	links TrioLinkList
}

// Defining a list type to manage uniqueness and ordering
type TrioDetailList []*TrioDetails

// All the initialized arrays used to navigate the switch between base trio index at base points
var AllBaseTrio [8]Trio
var ValidNextTrio [12][2]int
var AllMod4Permutations [12][4]int
var AllMod8Permutations [12][8]int

const(
	NbTrioDsIndex = 7
)

// All the possible Trio details used
var AllTrioDetails TrioDetailList


/***************************************************************/
// Util Functions
/***************************************************************/

func PosMod2(i uint64) uint64 {
	return i & 0x0000000000000001
}

func PosMod4(i uint64) uint64 {
	return i & 0x0000000000000003
}

func PosMod8(i uint64) uint64 {
	return i & 0x0000000000000007
}

// Test if in the base trio index i2 is pointing to the negative trio of i1 T[i1] == T'[i2]
func isPrime(i1, i2 int) bool {
	return i2-i1 == 4 || i2-i1 == -4
}

/***************************************************************/
// TrioLink Functions
/***************************************************************/

func makeTrioLink(a,b,c int) TrioLink {
	// The destination should be ordered by smaller first
	if c < b {
		return TrioLink{a,c,b,}
	}
	return TrioLink{a,b,c,}
}

func (tl TrioLink) sameBC() bool {
	return tl.b == tl.c
}

func (tl *TrioLink) String() string {
	return fmt.Sprintf("[%d %d %d]", tl.a, tl.b, tl.c)
}

/***************************************************************/
// TrioLinkList Functions
/***************************************************************/

func (l *TrioLinkList) exists(tl *TrioLink) bool {
	present := false
	for _, trL := range *l {
		if *trL == *tl {
			present = true
			break
		}
	}
	return present
}

func (l *TrioLinkList) addUnique(tl *TrioLink) bool {
	b := l.exists(tl)
	if !b {
		*l = append(*l, tl)
	}
	return b
}

func (l *TrioLinkList) addAll(l2 *TrioLinkList) {
	for _, tl := range *l2 {
		l.addUnique(tl)
	}
}

func (l TrioLinkList) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf( "%d : ", len(l)))
	for _, tl := range l {
		b.WriteString(tl.String())
		b.WriteString(" ")
	}
	return b.String()
}

func (l TrioLinkList) Len() int {
	return len(l)
}

func (l TrioLinkList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l TrioLinkList) Less(i, j int) bool {
	t1 := l[i]
	t2 := l[j]
	d := t1.a - t2.a
	if d != 0 {
		return d < 0
	}
	d = t1.b - t2.b
	if d != 0 {
		return d < 0
	}
	d = t1.c - t2.c
	if d != 0 {
		return d < 0
	}
	return false
}


/***************************************************************/
// TrioDetailsList Functions
/***************************************************************/

func (l *TrioDetailList) exists(tr *TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if trL.GetTrio() == tr.GetTrio() {
			present = true
			break
		}
	}
	return present
}

func (l *TrioDetailList) addUnique(td *TrioDetails) bool {
	b := l.exists(td)
	if !b {
		*l = append(*l, td)
	}
	return b
}

func (l *TrioDetailList) addWithLinks(td *TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if trL.GetTrio() == td.GetTrio() {
			trL.links.addAll(&td.links)
			present = true
			break
		}
	}
	if !present {
		*l = append(*l, td)
	}
	return present
}

func (l TrioDetailList) Len() int {
	return len(l)
}

func (l TrioDetailList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l TrioDetailList) Less(i, j int) bool {
	tr1 := l[i]
	tr2 := l[j]
	ds1Index := tr1.GetDSIndex()
	diffDS := ds1Index - tr2.GetDSIndex()

	// Order by ds index first
	if diffDS < 0 {
		return true
	} else if diffDS > 0 {
		return false
	} else {
		// Same ds index
		if ds1Index == 0 {
			// Base trio, keep order as defined with 0-4 prime -> 5-7
			var k, l int
			for bi, bt := range AllBaseTrio {
				if bt == tr1.GetTrio() {
					k = bi
				}
				if bt == tr2.GetTrio() {
					l = bi
				}
			}
			return k < l
		} else {
			// order by conn id, first ABS number, then pos > neg
			for k, cd1 := range tr1.conns {
				cd2 := tr2.conns[k]
				if cd1.GetIntId() != cd2.GetIntId() {
					return IsLessConnId(cd1, cd2)
				}
			}
		}
	}
	Log.Errorf("Should not get here for %v compare to %v", *tr1, *tr2)
	return false
}

/***************************************************************/
// Init Functions
// TODO: Find a better way than Init
/***************************************************************/

func init() {
	// Initial Trio 0
	AllBaseTrio[0] = MakeBaseConnectingVectorsTrio([3]Point{{1, 1, 0}, {-1, 0, -1}, {0, -1, 1}})
	for i := 1; i < 4; i++ {
		AllBaseTrio[i] = AllBaseTrio[i-1].PlusX()
	}
	// Initial Trio 0 prime
	for i := 0; i < 4; i++ {
		AllBaseTrio[i+4] = AllBaseTrio[i].Neg()
	}

	initValidTrios()
	initMod4Permutations()
	initMod8Permutations()
	initConnectionDetails()
	fillAllTrioDetails()
}

func initValidTrios() {
	// Valid next trio are all but prime
	idx := 0
	for i := 0; i < 4; i++ {
		for j := 4; j < 8; j++ {
			// j index cannot be the prime (neg) trio
			if !isPrime(i, j) {
				ValidNextTrio[idx] = [2]int{i, j}
				idx++
			}
		}
	}
}

func initMod4Permutations() {
	p := PermBuilder{4, 0, make([][]int, 12)}
	p.fill(0, make([]int, p.size))
	for pIdx := 0; pIdx < len(AllMod4Permutations); pIdx++ {
		for i := 0; i < 4; i++ {
			AllMod4Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

func initMod8Permutations() {
	p := PermBuilder{8, 0, make([][]int, 12)}
	// In 8 size permutation the first index always 0 since we use all the indexes
	first := make([]int, p.size)
	first[0] = 0
	p.fill(1, first)
	for pIdx := 0; pIdx < len(AllMod8Permutations); pIdx++ {
		for i := 0; i < 8; i++ {
			AllMod8Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

/***************************************************************/
// Trio Functions
/***************************************************************/
var reverse3Map = [3]int{2, 1, 0}

func MakeBaseConnectingVectorsTrio(points [3]Point) Trio {
	res := Trio{}
	// All points should be a connecting vector
	for _, p := range points {
		if !p.IsBaseConnectingVector() {
			Log.Error("Trying to create a base trio out of non base vector!", p)
			return res
		}
	}
	for _, p := range points {
		for i := 0; i < 3; i++ {
			if p[i] == 0 {
				res[reverse3Map[i]] = p
			}
		}
	}
	return res
}

func (t Trio) GetDSIndex() int {
	if t[0].DistanceSquared() == int64(1) {
		return 1
	} else {
		switch t[1].DistanceSquared() {
		case int64(2):
			return 0
		case int64(3):
			return 2
		case int64(5):
			return 3
		}
	}
	Log.Errorf("Did not find correct index for %v", t)
	return -1
}

func (t Trio) PlusX() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].PlusX(), t[1].PlusX(), t[2].PlusX()})
}

func (t Trio) Neg() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].Neg(), t[1].Neg(), t[2].Neg()})
}

// Return the 6 connections possible +X, -X, +Y, -Y, +Z, -Z vectors between 2 Trio
func GetNonBaseConnections(tA, tB Trio) [6]Point {
	res := [6]Point{}
	for _, aVec := range tA {
		switch aVec.X() {
		case 0:
			// Nothing connect
		case 1:
			// This is +X, find the -X on the other side
			bVec := tB.getMinusXVector()
			res[0] = XFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -X, find the +X on the other side
			bVec := tB.getPlusXVector()
			res[1] = XFirst.Neg().Add(bVec).Sub(aVec)
		}
		switch aVec.Y() {
		case 0:
			// Nothing connect
		case 1:
			// This is +Y, find the -Y on the other side
			bVec := tB.getMinusYVector()
			res[2] = YFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -Y, find the +Y on the other side
			bVec := tB.getPlusYVector()
			res[3] = YFirst.Neg().Add(bVec).Sub(aVec)
		}
		switch aVec.Z() {
		case 0:
			// Nothing connect
		case 1:
			// This is +Z, find the -Z on the other side
			bVec := tB.getMinusZVector()
			res[4] = ZFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -Z, find the +Z on the other side
			bVec := tB.getPlusZVector()
			res[5] = ZFirst.Neg().Add(bVec).Sub(aVec)
		}
	}
	return res
}

func (t Trio) getPlusXVector() Point {
	for _, vec := range t {
		if vec.X() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a +X vector")
	return Origin
}

func (t Trio) getMinusXVector() Point {
	for _, vec := range t {
		if vec.X() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a -X vector")
	return Origin
}

func (t Trio) getPlusYVector() Point {
	for _, vec := range t {
		if vec.Y() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a +Y vector")
	return Origin
}

func (t Trio) getMinusYVector() Point {
	for _, vec := range t {
		if vec.Y() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a -Y vector")
	return Origin
}

func (t Trio) getPlusZVector() Point {
	for _, vec := range t {
		if vec.Z() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a +Z vector")
	return Origin
}

func (t Trio) getMinusZVector() Point {
	for _, vec := range t {
		if vec.Z() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all trio there should be a -Z vector")
	return Origin
}

/***************************************************************/
// TrioDetails Functions
/***************************************************************/
var count = 0

func MakeTrioDetails(points ...Point) *TrioDetails {
	// All points should be a connection details
	cds := make([]*ConnectionDetails, 3)
	for i, p := range points {
		cd, ok := AllConnectionsPossible[p]
		if !ok {
			Log.Fatalf("trying to create trio with vector not a connection %v", p)
		} else {
			cds[i] = cd
		}
	}
	// Order based on connection details index, and if same index Pos > Neg
	sort.Slice(cds, func(i, j int) bool {
		absDiff := cds[i].GetPosIntId() - cds[j].GetPosIntId()
		if absDiff == 0 {
			return cds[i].Id > 0
		}
		return absDiff < 0
	})
	res := TrioDetails{}
	res.id = -1
	for i, cd := range cds {
		res.conns[i] = cd
	}
	if Log.Level <= m3util.TRACE {
		fmt.Println(count, res.conns[0].GetName(), res.conns[1].GetName(), res.conns[2].GetName())
		count++
	}
	return &res
}

func (td *TrioDetails) String() string {
	return fmt.Sprintf("T%02d: (%s, %s, %s) l=%3d", td.id, td.conns[0].GetName(), td.conns[1].GetName(), td.conns[2].GetName(), len(td.links))
}

func (td *TrioDetails) GetTrio() Trio {
	return Trio{td.conns[0].Vector, td.conns[1].Vector, td.conns[2].Vector}
}

func (td *TrioDetails) GetDSIndex() int {
	if td.conns[0].DistanceSquared() == int64(1) {
		switch td.conns[1].DistanceSquared() {
		case int64(1):
			return 1
		case int64(2):
			switch td.conns[2].DistanceSquared() {
			case int64(3):
				return 2
			case int64(5):
				return 3
			}
		}
	} else {
		switch td.conns[1].DistanceSquared() {
		case int64(2):
			return 0
		case int64(3):
			switch td.conns[2].DistanceSquared() {
			case int64(3):
				return 4
			case int64(5):
				return 5
			}
		case int64(5):
			return 6
		}
	}
	Log.Errorf("Did not find correct index for %v", *td)
	return -1
}

func fillAllTrioDetails() {
	allTrioLinks := TrioLinkList{}
	allTDSlice := TrioDetailList{}
	// All base trio first
	for i, tr := range AllBaseTrio {
		td := MakeTrioDetails(tr[0], tr[1], tr[2])
		td.id = i
		allTDSlice.addUnique(td)
	}
	// Going through all Trio and all combination of Trio, to find middle points and create new Trios
	for a, tA := range AllBaseTrio {
		for b, tB := range AllBaseTrio {
			for c, tC := range AllBaseTrio {
				thisTrio := makeTrioLink(a,b,c)
				alreadyDone := allTrioLinks.addUnique(&thisTrio)
				if !alreadyDone {
					for _, nextTrio := range getNextTriosDetails(tA, tB, tC) {
						nextTrio.links.addUnique(&thisTrio)
						allTDSlice.addWithLinks(nextTrio)
					}
				}
			}
		}
	}

	sort.Sort(allTDSlice)

	for i, td := range allTDSlice {
		if td.id != -1 && td.id != i {
			Log.Fatalf("incorrect Id for trio details %v at %d", *td, i)
		}
		td.id = i
		// Order the links array
		sort.Sort(td.links)
	}
	AllTrioDetails = allTDSlice
}

// Return the new Trio out of Origin + tA (with next tB or tB/tC)
func getNextTriosDetails(tA, tB, tC Trio) []*TrioDetails {
	// 0 z=0 for first element, x connector, y connector
	// 1 y=0 for first element, x connector, z connector
	// 2 x=0 for first element, y connector, z connector
	res := TrioDetailList{}

	sameBC := tB == tC
	noZ := tA[0]
	var xConnB, yConnB, zConnB Point
	var xConnC, yConnC, zConnC Point
	if noZ.X() > 0 {
		xConnB = MakeVector(noZ, XFirst.Add(tB.getMinusXVector()))
		if !sameBC {
			xConnC = MakeVector(noZ, XFirst.Add(tC.getMinusXVector()))
		}
	} else {
		xConnB = MakeVector(noZ, XFirst.Neg().Add(tB.getPlusXVector()))
		if !sameBC {
			xConnC = MakeVector(noZ, XFirst.Neg().Add(tC.getPlusXVector()))
		}
	}
	if noZ.Y() > 0 {
		yConnB = MakeVector(noZ, YFirst.Add(tB.getMinusYVector()))
		if !sameBC {
			yConnC = MakeVector(noZ, YFirst.Add(tC.getMinusYVector()))
		}
	} else {
		yConnB = MakeVector(noZ, YFirst.Neg().Add(tB.getPlusYVector()))
		if !sameBC {
			yConnC = MakeVector(noZ, YFirst.Neg().Add(tC.getPlusYVector()))
		}
	}
	if sameBC {
		res.addUnique(MakeTrioDetails(noZ.Neg(), xConnB, yConnB))
	} else {
		res.addUnique(MakeTrioDetails(noZ.Neg(), xConnB, yConnC))
		res.addUnique(MakeTrioDetails(noZ.Neg(), xConnC, yConnB))
	}

	noY := tA[1]
	if noY.X() > 0 {
		xConnB = MakeVector(noY, XFirst.Add(tB.getMinusXVector()))
		if !sameBC {
			xConnC = MakeVector(noY, XFirst.Add(tC.getMinusXVector()))
		}
	} else {
		xConnB = MakeVector(noY, XFirst.Neg().Add(tB.getPlusXVector()))
		if !sameBC {
			xConnC = MakeVector(noY, XFirst.Neg().Add(tC.getPlusXVector()))
		}
	}
	if noY.Z() > 0 {
		zConnB = MakeVector(noY, ZFirst.Add(tB.getMinusZVector()))
		if !sameBC {
			zConnC = MakeVector(noY, ZFirst.Add(tC.getMinusZVector()))
		}
	} else {
		zConnB = MakeVector(noY, ZFirst.Neg().Add(tB.getPlusZVector()))
		if !sameBC {
			zConnC = MakeVector(noY, ZFirst.Neg().Add(tC.getPlusZVector()))
		}
	}
	if sameBC {
		res.addUnique(MakeTrioDetails(noY.Neg(), xConnB, zConnB))
	} else {
		res.addUnique(MakeTrioDetails(noY.Neg(), xConnB, zConnC))
		res.addUnique(MakeTrioDetails(noY.Neg(), xConnC, zConnB))
	}

	noX := tA[2]
	if noX.Y() > 0 {
		yConnB = MakeVector(noX, YFirst.Add(tB.getMinusYVector()))
		if !sameBC {
			yConnC = MakeVector(noX, YFirst.Add(tC.getMinusYVector()))
		}
	} else {
		yConnB = MakeVector(noX, YFirst.Neg().Add(tB.getPlusYVector()))
		if !sameBC {
			yConnC = MakeVector(noX, YFirst.Neg().Add(tC.getPlusYVector()))
		}
	}
	if noX.Z() > 0 {
		zConnB = MakeVector(noX, ZFirst.Add(tB.getMinusZVector()))
		if !sameBC {
			zConnC = MakeVector(noX, ZFirst.Add(tC.getMinusZVector()))
		}
	} else {
		zConnB = MakeVector(noX, ZFirst.Neg().Add(tB.getPlusZVector()))
		if !sameBC {
			zConnC = MakeVector(noX, ZFirst.Neg().Add(tC.getPlusZVector()))
		}
	}
	if sameBC {
		res.addUnique(MakeTrioDetails(noX.Neg(), yConnB, zConnB))
	} else {
		res.addUnique(MakeTrioDetails(noX.Neg(), yConnB, zConnC))
		res.addUnique(MakeTrioDetails(noX.Neg(), yConnC, zConnB))
	}

	return res
}
