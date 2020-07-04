package m3server

import (
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"sort"
)

// trio of connection vectors from any m3point.Point using connections only.
// This type is used to calculate the TrioDetails and ignored after.
// All of the rest of the code use mainly the trio index value.
type trio [3]m3point.Point

// Defining a list type to manage uniqueness and ordering
type TrioDetailList []*m3point.TrioDetails
type ByConnVector []*m3point.ConnectionDetails
type ByConnId []*m3point.ConnectionDetails

var reverse3Map = [3]m3point.TrioIndex{2, 1, 0}

// All the initialized arrays used to navigate the switch between base trio index at base m3point.Points
var allBaseTrio [8]trio

// TODO: Use the arrays in BasePointPackData
var validNextTrio [12][2]m3point.TrioIndex
var AllMod4Permutations [12][4]m3point.TrioIndex
var AllMod8Permutations [12][8]m3point.TrioIndex

// Dummy trace counter for debugging recursive methods
var _traceCounter = 0

var Origin = m3point.Point{0, 0, 0}
var XFirst = m3point.Point{m3point.THREE, 0, 0}
var YFirst = m3point.Point{0, m3point.THREE, 0}
var ZFirst = m3point.Point{0, 0, m3point.THREE}

/***************************************************************/
// Init Functions
/***************************************************************/

func init() {
	// Initial trio 0
	allBaseTrio[0] = makeBaseConnectingVectorsTrio([3]m3point.Point{{1, 1, 0}, {-1, 0, -1}, {0, -1, 1}})
	for i := 1; i < 4; i++ {
		allBaseTrio[i] = allBaseTrio[i-1].RotPlusX()
	}
	// Initial trio 0 prime
	for i := 0; i < 4; i++ {
		allBaseTrio[i+4] = allBaseTrio[i].Neg()
	}

	initValidTrios()
	initMod4Permutations()
	initMod8Permutations()
}

// Test if in the base trio index i2 is pointing to the negative trio of i1 T[i1] == T'[i2]
func isPrime(i1, i2 m3point.TrioIndex) bool {
	if !i1.IsBaseTrio() || !i2.IsBaseTrio() {
		Log.Errorf("cannot compare prime for non base trio index %d %d")
		return false
	}
	if i1 > i2 {
		return i1 == i2+4
	}
	if i2 > i1 {
		return i2 == i1+4
	}
	return false
}

func initValidTrios() {
	// Valid next trio are all but prime
	idx := 0
	for i := m3point.TrioIndex(0); i < 4; i++ {
		for j := m3point.TrioIndex(4); j < 8; j++ {
			// j index cannot be the prime (neg) trio
			if !isPrime(i, j) {
				validNextTrio[idx] = [2]m3point.TrioIndex{i, j}
				idx++
			}
		}
	}
}

func initMod4Permutations() {
	p := TrioIndexPermBuilder{4, 0, make([][]m3point.TrioIndex, 12)}
	p.fill(0, make([]m3point.TrioIndex, p.size))
	for pIdx := 0; pIdx < len(AllMod4Permutations); pIdx++ {
		for i := 0; i < 4; i++ {
			AllMod4Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

func initMod8Permutations() {
	p := TrioIndexPermBuilder{8, 0, make([][]m3point.TrioIndex, 12)}
	// In 8 size permutation the first index always 0 since we use all the indexes
	first := make([]m3point.TrioIndex, p.size)
	first[0] = m3point.TrioIndex(0)
	p.fill(1, first)
	for pIdx := 0; pIdx < len(AllMod8Permutations); pIdx++ {
		for i := 0; i < 8; i++ {
			AllMod8Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

/***************************************************************/
// ByConnVector functions
/***************************************************************/

func (cds ByConnVector) Len() int      { return len(cds) }
func (cds ByConnVector) Swap(i, j int) { cds[i], cds[j] = cds[j], cds[i] }
func (cds ByConnVector) Less(i, j int) bool {
	cd1 := cds[i]
	cd2 := cds[j]
	dsDiff := cd1.ConnDS - cd2.ConnDS
	if dsDiff == 0 {
		// X < Y < Z
		for c := 0; c < 3; c++ {
			d := m3point.AbsDIntFromC(cd1.Vector[c]) - m3point.AbsDIntFromC(cd2.Vector[c])
			if d != 0 {
				return d > 0
			}
		}
		// All abs value equal the first coord that is positive is less
		for c := 0; c < 3; c++ {
			d := cd1.Vector[c] - cd2.Vector[c]
			if d != 0 {
				return d > 0
			}
		}
	}
	return dsDiff < 0
}

func (cds ByConnId) Len() int      { return len(cds) }
func (cds ByConnId) Swap(i, j int) { cds[i], cds[j] = cds[j], cds[i] }
func (cds ByConnId) Less(i, j int) bool {
	return IsLessConnId(cds[i], cds[j])
}

func IsLessConnId(cd1, cd2 *m3point.ConnectionDetails) bool {
	absDiff := cd1.GetPosId() - cd2.GetPosId()
	if absDiff < 0 {
		return true
	} else if absDiff > 0 {
		return false
	} else {
		return cd1.Id > 0
	}
}

/***************************************************************/
// trio Functions
/***************************************************************/

func makeBaseConnectingVectorsTrio(points [3]m3point.Point) trio {
	res := trio{}
	// All m3point.Points should be a connecting vector
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

func (t trio) GetDSIndex() int {
	if t[0].DistanceSquared() == m3point.DInt(1) {
		return 1
	} else {
		switch t[1].DistanceSquared() {
		case m3point.DInt(2):
			return 0
		case m3point.DInt(3):
			return 2
		case m3point.DInt(5):
			return 3
		}
	}
	Log.Errorf("Did not find correct index for %v", t)
	return -1
}

func (t trio) RotPlusX() trio {
	return makeBaseConnectingVectorsTrio([3]m3point.Point{t[0].RotPlusX(), t[1].RotPlusX(), t[2].RotPlusX()})
}

func (t trio) Neg() trio {
	return makeBaseConnectingVectorsTrio([3]m3point.Point{t[0].Neg(), t[1].Neg(), t[2].Neg()})
}

// Return the 6 connections possible +X, -X, +Y, -Y, +Z, -Z vectors between 2 trio
func GetNonBaseConnections(tA, tB trio) [6]m3point.Point {
	res := [6]m3point.Point{}
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

func (t trio) getPlusXVector() m3point.Point {
	for _, vec := range t {
		if vec.X() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a +X vector")
	return Origin
}

func (t trio) getMinusXVector() m3point.Point {
	for _, vec := range t {
		if vec.X() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a -X vector")
	return Origin
}

func (t trio) getPlusYVector() m3point.Point {
	for _, vec := range t {
		if vec.Y() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a +Y vector")
	return Origin
}

func (t trio) getMinusYVector() m3point.Point {
	for _, vec := range t {
		if vec.Y() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a -Y vector")
	return Origin
}

func (t trio) getPlusZVector() m3point.Point {
	for _, vec := range t {
		if vec.Z() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a +Z vector")
	return Origin
}

func (t trio) getMinusZVector() m3point.Point {
	for _, vec := range t {
		if vec.Z() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a -Z vector")
	return Origin
}

/***************************************************************/
// Calculate Connections and trio Functions
/***************************************************************/

func (ppd *PointPackData) calculateConnectionDetails() ([]*m3point.ConnectionDetails, map[m3point.Point]*m3point.ConnectionDetails) {
	connMap := make(map[m3point.Point]*m3point.ConnectionDetails)
	// Going through all trio and all combination of trio, to aggregate connection details
	for _, tr := range allBaseTrio {
		for _, vec := range tr {
			addConnDetail(&connMap, vec)
		}
		for _, tB := range allBaseTrio {
			connectingVectors := GetNonBaseConnections(tr, tB)
			for _, conn := range connectingVectors {
				addConnDetail(&connMap, conn)
			}
		}
	}
	Log.Debug("Number of connection details created", len(connMap))
	nbConnDetails := m3point.ConnectionId(len(connMap) / 2)

	// Reordering connection details number by size, and x, y, z
	res := make([]*m3point.ConnectionDetails, len(connMap))
	idx := 0
	for _, cd := range connMap {
		res[idx] = cd
		idx++
	}
	sort.Sort(ByConnVector(res))

	currentConnNumber := m3point.ConnectionId(1)
	for _, cd := range res {
		if cd.Id == 0 {
			vec1 := cd.Vector
			vec2 := vec1.Neg()
			var posVec, negVec m3point.Point
			// first one with non 0 pos coord
			for _, c := range vec1 {
				if c > 0 {
					posVec = vec1
					negVec = vec2
					break
				} else if c < 0 {
					posVec = vec2
					negVec = vec1
					break
				}
			}
			posCD := connMap[posVec]
			posCD.Id = currentConnNumber
			negCD := connMap[negVec]
			negCD.Id = -currentConnNumber
			currentConnNumber++
		}
	}
	sort.Sort(ByConnId(res))

	lastId := res[len(res)-2].GetId()
	if lastId != nbConnDetails {
		Log.Errorf("Calculating Connection details failed since %d != %d", lastId, nbConnDetails)
	}
	return res, connMap
}

func addConnDetail(connMap *map[m3point.Point]*m3point.ConnectionDetails, connVector m3point.Point) {
	ds := connVector.DistanceSquared()
	if ds == 0 {
		Log.Fatalf("zero vector cannot be a connection")
	}
	if !(ds == 1 || ds == 2 || ds == 3 || ds == 5) {
		Log.Fatalf("vector %v of ds=%d cannot be a connection", connVector, ds)
	}
	_, ok := (*connMap)[connVector]
	if !ok {
		// Add both pos and neg
		posVec := connVector
		negVec := connVector.Neg()
		posConnDetails := m3point.ConnectionDetails{Id: 0, Vector: posVec, ConnDS: ds}
		negConnDetails := m3point.ConnectionDetails{Id: 0, Vector: negVec, ConnDS: ds}
		(*connMap)[posVec] = &posConnDetails
		(*connMap)[negVec] = &negConnDetails
	}
}

func (ppd *PointPackData) calculateAllTrioDetails() TrioDetailList {
	res := TrioDetailList(make([]*m3point.TrioDetails, 0, 200))
	// All base trio first
	for i, tr := range allBaseTrio {
		td := ppd.makeTrioDetails(tr[0], tr[1], tr[2])
		td.Id = m3point.TrioIndex(i)
		res.addUnique(td)
	}

	// Going through all trio and all combination of trio, to find middle m3point.Points and create new Trios
	for _, tA := range allBaseTrio {
		for _, tB := range allBaseTrio {
			for _, tC := range allBaseTrio {
				for _, nextTrio := range ppd.getNextTriosDetails(tA, tB, tC) {
					res.addUnique(nextTrio)
				}
			}
		}
	}

	sort.Sort(res)

	// Process all the trio details now that order final
	for i, td := range res {
		trIdx := m3point.TrioIndex(i)
		// For all base trio different process
		if i < len(allBaseTrio) {
			// The id should already be set correctly
			if td.Id != trIdx {
				Log.Fatalf("incorrect Id for base trio details %v at %d", *td, i)
			}
		} else {
			// The id should not have been set. Adding it now
			if td.Id != m3point.NilTrioIndex {
				Log.Fatalf("incorrect Id for non base trio details %v at %d", *td, i)
			}
			td.Id = trIdx
		}
	}
	return res
}

func (ppd *PointPackData) makeTrioDetails(points ...m3point.Point) *m3point.TrioDetails {
	// All m3point.Points should be a connection details
	cds := make([]*m3point.ConnectionDetails, 3)
	for i, p := range points {
		cd := ppd.GetConnDetailsByVector(p)
		if cd == nil {
			Log.Fatalf("trying to create trio with vector not a connection %v", p)
		}
		cds[i] = cd
	}
	// Order based on connection details index, and if same index Pos > Neg
	sort.Slice(cds, func(i, j int) bool {
		absDiff := cds[i].GetPosId() - cds[j].GetPosId()
		if absDiff == 0 {
			return cds[i].Id > 0
		}
		return absDiff < 0
	})
	res := m3point.TrioDetails{}
	res.Id = m3point.NilTrioIndex
	for i, cd := range cds {
		res.Conns[i] = cd
	}
	if Log.IsTrace() {
		fmt.Println(_traceCounter, res.Conns[0].String(), res.Conns[1].String(), res.Conns[2].String())
		_traceCounter++
	}
	return &res
}

func MakeVector(p1, p2 m3point.Point) m3point.Point {
	return p2.Sub(p1)
}

// Return the new trio out of Origin + tA (with next tB or tB/tC)
func (ppd *PointPackData) getNextTriosDetails(tA, tB, tC trio) []*m3point.TrioDetails {
	// 0 z=0 for first element, x connector, y connector
	// 1 y=0 for first element, x connector, z connector
	// 2 x=0 for first element, y connector, z connector
	res := TrioDetailList{}

	sameBC := tB == tC
	noZ := tA[0]
	var xConnB, yConnB, zConnB m3point.Point
	var xConnC, yConnC, zConnC m3point.Point
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
		res.addUnique(ppd.makeTrioDetails(noZ.Neg(), xConnB, yConnB))
	} else {
		res.addUnique(ppd.makeTrioDetails(noZ.Neg(), xConnB, yConnC))
		res.addUnique(ppd.makeTrioDetails(noZ.Neg(), xConnC, yConnB))
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
		res.addUnique(ppd.makeTrioDetails(noY.Neg(), xConnB, zConnB))
	} else {
		res.addUnique(ppd.makeTrioDetails(noY.Neg(), xConnB, zConnC))
		res.addUnique(ppd.makeTrioDetails(noY.Neg(), xConnC, zConnB))
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
		res.addUnique(ppd.makeTrioDetails(noX.Neg(), yConnB, zConnB))
	} else {
		res.addUnique(ppd.makeTrioDetails(noX.Neg(), yConnB, zConnC))
		res.addUnique(ppd.makeTrioDetails(noX.Neg(), yConnC, zConnB))
	}

	return res
}

/***************************************************************/
// TrioDetailsList Functions
/***************************************************************/

func (l *TrioDetailList) IdList() []m3point.TrioIndex {
	res := make([]m3point.TrioIndex, len(*l))
	for i, td := range *l {
		res[i] = td.GetId()
	}
	return res
}

func convertToTrio(td *m3point.TrioDetails) trio {
	return trio{td.Conns[0].Vector, td.Conns[1].Vector, td.Conns[2].Vector}
}

func (l *TrioDetailList) ExistsByTrio(tr *m3point.TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if convertToTrio(trL) == convertToTrio(tr) {
			present = true
			break
		}
	}
	return present
}

func (l *TrioDetailList) ExistsById(tr *m3point.TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if trL.Id == tr.Id {
			present = true
			break
		}
	}
	return present
}

func (l *TrioDetailList) addUnique(td *m3point.TrioDetails) bool {
	b := l.ExistsByTrio(td)
	if !b {
		*l = append(*l, td)
	}
	return b
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
			for bi, bt := range allBaseTrio {
				if bt == convertToTrio(tr1) {
					k = bi
				}
				if bt == convertToTrio(tr2) {
					l = bi
				}
			}
			return k < l
		} else {
			// order by conn id, first ABS number, then pos > neg
			for k, cd1 := range tr1.Conns {
				cd2 := tr2.Conns[k]
				if cd1.GetId() != cd2.GetId() {
					return IsLessConnId(cd1, cd2)
				}
			}
		}
	}
	Log.Errorf("Should not get here for %v compare to %v", *tr1, *tr2)
	return false
}

/***************************************************************/
// trio Details Load and Save
/***************************************************************/

func (ppd *PointPackData) loadTrioDetails() TrioDetailList {
	te, rows := ppd.Env.SelectAllForLoad(TrioDetailsTable)

	res := TrioDetailList(make([]*m3point.TrioDetails, 0, te.TableDef.ExpectedCount))

	for rows.Next() {
		td := m3point.TrioDetails{}
		connIds := [3]m3point.ConnectionId{}
		err := rows.Scan(&td.Id, &connIds[0], &connIds[1], &connIds[2])
		if err != nil {
			Log.Errorf("failed to load trio details line %d", len(res))
		} else {
			for i, cId := range connIds {
				td.Conns[i] = ppd.GetConnDetailsById(cId)
			}
			res = append(res, &td)
		}
	}
	return res
}

func (ppd *PointPackData) saveAllTrioDetails() (int, error) {
	te, inserted, toFill, err := ppd.Env.GetForSaveAll(TrioDetailsTable)
	if te == nil {
		return 0, err
	}

	if toFill {
		trios := ppd.calculateAllTrioDetails()
		if Log.IsDebug() {
			Log.Debugf("Populating table %s with %d elements", te.TableDef.Name, len(trios))
		}
		for _, td := range trios {
			err := te.Insert(td.Id, td.Conns[0].Id, td.Conns[1].Id, td.Conns[2].Id)
			if err != nil {
				Log.Error(err)
			} else {
				inserted++
			}
		}
	}
	return inserted, nil
}
