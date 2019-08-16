package m3point

import (
	"fmt"
)

/***************************************************************/
// Type declaration
/***************************************************************/

type UnitDirection int

type ConnectionId int8

type ConnectionDetails struct {
	Id     ConnectionId
	Vector Point
	ConnDS DInt
}

type ByConnVector []*ConnectionDetails
type ByConnId []*ConnectionDetails

/***************************************************************/
// Global fields declaration
/***************************************************************/

const (
	PlusX UnitDirection = iota
	MinusX
	PlusY
	MinusY
	PlusZ
	MinusZ
)

var NilConnectionId = ConnectionId(0)
var EmptyConnDetails = ConnectionDetails{NilConnectionId, Origin, 0,}

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
			d := AbsDIntFromC(cd1.Vector[c]) - AbsDIntFromC(cd2.Vector[c])
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

func IsLessConnId(cd1, cd2 *ConnectionDetails) bool {
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
// ConnectionId Functions
/***************************************************************/
func (connId ConnectionId) IsValid() bool {
	return connId != NilConnectionId
}

func (connId ConnectionId) GetNegId() ConnectionId {
	return -connId
}

func (connId ConnectionId) GetPosConnectionId() ConnectionId {
	if connId < 0 {
		return -connId
	}
	return connId
}

func (connId ConnectionId) IsBaseConnection() bool {
	posConnId := connId.GetPosConnectionId()
	return posConnId >= 4 && posConnId <= 9
}

func (connId ConnectionId) String() string {
	if connId < 0 {
		return fmt.Sprintf("CN%02d", -connId)
	} else {
		return fmt.Sprintf("CP%02d", connId)
	}
}

/***************************************************************/
// UnitDirection Functions
/***************************************************************/

var NegXFirst = XFirst.Neg()
var NegYFirst = YFirst.Neg()
var NegZFirst = ZFirst.Neg()

func (ud UnitDirection) GetOpposite() UnitDirection {
	switch ud {
	case PlusX:
		return MinusX
	case MinusX:
		return PlusX
	case PlusY:
		return MinusY
	case MinusY:
		return PlusY
	case PlusZ:
		return MinusZ
	case MinusZ:
		return PlusZ
	}
	Log.Fatalf("Impossible! Did not find %d unit direction", ud)
	return UnitDirection(-1)
}

func (ud UnitDirection) GetFirstPoint() Point {
	switch ud {
	case PlusX:
		return XFirst
	case MinusX:
		return NegXFirst
	case PlusY:
		return YFirst
	case MinusY:
		return NegYFirst
	case PlusZ:
		return ZFirst
	case MinusZ:
		return NegZFirst
	}
	Log.Fatalf("Impossible! Did not find %d unit direction", ud)
	return Origin
}

/***************************************************************/
// ConnectionDetails Functions
/***************************************************************/

func (cd *ConnectionDetails) IsValid() bool {
	return cd.Id.IsValid()
}

func (cd *ConnectionDetails) GetId() ConnectionId {
	return cd.Id
}

func (cd *ConnectionDetails) GetNegId() ConnectionId {
	return cd.Id.GetNegId()
}

func (cd *ConnectionDetails) GetPosId() ConnectionId {
	return cd.Id.GetPosConnectionId()
}

func (cd *ConnectionDetails) IsBaseConnection() bool {
	return cd.Id.IsBaseConnection()
}

func (cd *ConnectionDetails) GetDirections() [2]UnitDirection {
	if !cd.IsBaseConnection() {
		Log.Fatalf("cannot extract unit directions on non base connection %s", cd.String())
	}
	idx := 0
	res := [2]UnitDirection{}
	cVec := cd.Vector
	switch cVec.X() {
	case 0:
		// Nothing connect
	case 1:
		// Going +X
		res[idx] = PlusX
		idx++
	case -1:
		// Going -X
		res[idx] = MinusX
		idx++
	}
	switch cVec.Y() {
	case 0:
		// Nothing connect
	case 1:
		// Going +Y
		res[idx] = PlusY
		idx++
	case -1:
		// Going -Y
		res[idx] = MinusY
		idx++
	}
	switch cVec.Z() {
	case 0:
		// Nothing connect
	case 1:
		// Going +Z
		res[idx] = PlusZ
		idx++
	case -1:
		// Going -Z
		res[idx] = MinusZ
		idx++
	}
	return res
}

func (cd *ConnectionDetails) DistanceSquared() DInt {
	absId := cd.GetPosId()
	if absId <= 3 {
		return DInt(1)
	} else if absId <= 9 {
		return DInt(2)
	} else if absId <= 13 {
		return DInt(3)
	} else if absId <= 25 {
		return DInt(5)
	}
	Log.Fatalf("abs Id of connection details %v invalid", cd)
	return DInt(0)
}

func (cd *ConnectionDetails) String() string {
	return cd.Id.String()
}

/***************************************************************/
// PointPackData Functions for ConnectionDetails
/***************************************************************/

func (ppd *PointPackData) GetMaxConnId() ConnectionId {
	ppd.checkConnInitialized()
	// The pos conn Id of the last one
	return ppd.allConnections[len(ppd.allConnections)-1].GetPosId()
}

func (ppd *PointPackData) GetConnDetailsById(id ConnectionId) *ConnectionDetails {
	ppd.checkConnInitialized()
	if id > 0 {
		return ppd.allConnections[2*id-2]
	} else {
		return ppd.allConnections[-2*id-1]
	}
}

func (ppd *PointPackData) GetConnDetailsByPoints(p1, p2 Point) *ConnectionDetails {
	return ppd.getConnDetailsByVector(MakeVector(p1, p2))
}

func (ppd *PointPackData) getAllConnDetailsByVector() map[Point]*ConnectionDetails {
	ppd.checkConnInitialized()
	return ppd.allConnectionsByVector
}

func (ppd *PointPackData) getConnDetailsByVector(vector Point) *ConnectionDetails {
	ppd.checkConnInitialized()
	cd, ok := ppd.allConnectionsByVector[vector]
	if !ok {
		Log.Error("Vector", vector, "is not a known connection details")
		return &EmptyConnDetails
	}
	return cd
}
