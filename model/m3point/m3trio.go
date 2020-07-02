package m3point

import (
	"fmt"
)

/***************************************************************/
// Type declaration
/***************************************************************/

// The unique well defined number of a Trio used on a connected point which we call later a Node
type TrioIndex uint8

// A bigger struct than trio to keep more info on how points grow from a trio index
type TrioDetails struct {
	Id    TrioIndex
	Conns [3]*ConnectionDetails
}

/***************************************************************/
// Global fields declaration
/***************************************************************/

const (
	NbTrioDsIndex = 7
	NilTrioIndex  = TrioIndex(255)
)

/***************************************************************/
// TrioIndex Functions
/***************************************************************/

func (trIdx TrioIndex) IsBaseTrio() bool {
	return trIdx < 8
}

func (trIdx TrioIndex) String() string {
	return fmt.Sprintf("T%03d", trIdx)
}

/***************************************************************/
// GetTrioDetails Functions
/***************************************************************/

func (td *TrioDetails) String() string {
	return fmt.Sprintf("T%02d: (%s, %s, %s)", td.Id, td.Conns[0].String(), td.Conns[1].String(), td.Conns[2].String())
}

func (td *TrioDetails) GetConnectionIdx(connId ConnectionId) int {
	for idx, c := range td.Conns {
		if c.Id == connId {
			return idx
		}
	}
	return -1
}

func (td *TrioDetails) HasConnection(connId ConnectionId) bool {
	for _, c := range td.Conns {
		if c.Id == connId {
			return true
		}
	}
	return false
}

// The passed connId is where come from so is neg in here
func (td *TrioDetails) OtherConnectionsFrom(connId ConnectionId) [2]*ConnectionDetails {
	res := [2]*ConnectionDetails{nil, nil}
	idx := 0

	if td.HasConnection(connId) {
		for _, c := range td.Conns {
			if c.Id != connId {
				res[idx] = c
				idx++
			}
		}
	} else {
		Log.Errorf("connection %s is not part of %s and cannot return other connections", connId.String(), td.String())
	}

	return res
}

func (td *TrioDetails) LastOtherConnection(cIds ...ConnectionId) *ConnectionDetails {
	if Log.DoAssert() {
		if len(cIds) != 2 {
			Log.Errorf("calling LastOtherConnection on %s not using 2 other connections %v", td.String(), cIds)
		}
		if cIds[0] == cIds[1] {
			Log.Errorf("calling LastOtherConnection on %s with 2 identical connections %v", td.String(), cIds)
		}
		for _, cId := range cIds {
			if !td.HasConnection(cId) {
				Log.Errorf("calling LastOtherConnection on %s with connections %v and %s is not in trio", td.String(), cIds, cId.String())
			}
		}
	}
	for _, c := range td.Conns {
		found := false
		for _, cId := range cIds {
			if c.Id == cId {
				found = true
			}
		}
		if !found {
			return c
		}
	}
	Log.Errorf("calling LastOtherConnection on %s with connections %v and nothing found in trio!", td.String(), cIds)
	return nil
}

func (td *TrioDetails) HasConnections(cIds ...ConnectionId) bool {
	for _, cId := range cIds {
		if !td.HasConnection(cId) {
			return false
		}
	}
	return true
}

func (td *TrioDetails) GetConnections() [3]*ConnectionDetails {
	return td.Conns
}

func (td *TrioDetails) GetId() TrioIndex {
	return td.Id
}

func (td *TrioDetails) IsBaseTrio() bool {
	return td.Id < 8
}

func (td *TrioDetails) findConn(vecName string, toFind ...ConnectionId) *ConnectionDetails {
	if !td.IsBaseTrio() {
		Log.Errorf("cannot look for %s conn on non base trio %s", vecName, td.String())
		return nil
	}
	if Log.DoAssert() {
		// verify only one found
		count := 0
		var res *ConnectionDetails
		for _, c := range td.Conns {
			for _, ct := range toFind {
				if c.Id == ct {
					res = c
					count++
				}
			}
		}
		if count == 0 {
			Log.Errorf("Impossible! Did not find %s vector using %v in base trio %s", vecName, toFind, td.String())
			return nil
		} else if count > 1 {
			Log.Errorf("Found %d which is more than one %s vector using %v in base trio %s", count, vecName, toFind, td.String())
			return nil
		}
		return res
	} else {
		for _, c := range td.Conns {
			for _, ct := range toFind {
				if c.Id == ct {
					return c
				}
			}
		}
		Log.Errorf("Impossible! Did not find %s vector using %v in base trio %s", vecName, toFind, td.String())
		return nil
	}
}

func (td *TrioDetails) getConn(ud UnitDirection) *ConnectionDetails {
	switch ud {
	case PlusX:
		return td.GetPlusXConn()
	case MinusX:
		return td.GetMinusXConn()
	case PlusY:
		return td.GetPlusYConn()
	case MinusY:
		return td.GetMinusYConn()
	case PlusZ:
		return td.GetPlusZConn()
	case MinusZ:
		return td.GetMinusZConn()
	}
	Log.Fatalf("Impossible! Did not find %d unit direction", ud)
	return nil
}

func (td *TrioDetails) GetOppositeConn(ud UnitDirection) *ConnectionDetails {
	return td.getConn(ud.GetOpposite())
}

func (td *TrioDetails) GetPlusXConn() *ConnectionDetails {
	return td.findConn("+X", 4, 5, 6, 7)
}

func (td *TrioDetails) GetMinusXConn() *ConnectionDetails {
	return td.findConn("-X", -4, -5, -6, -7)
}

func (td *TrioDetails) GetPlusYConn() *ConnectionDetails {
	return td.findConn("+Y", 4, -5, 8, 9)
}

func (td *TrioDetails) GetMinusYConn() *ConnectionDetails {
	return td.findConn("-Y", -4, 5, -8, -9)
}

func (td *TrioDetails) GetPlusZConn() *ConnectionDetails {
	return td.findConn("+Z", 6, -7, 8, -9)
}

func (td *TrioDetails) GetMinusZConn() *ConnectionDetails {
	return td.findConn("-Z", -6, 7, -8, 9)
}

func (td *TrioDetails) GetDSIndex() int {
	if td.Conns[0].DistanceSquared() == DInt(1) {
		switch td.Conns[1].DistanceSquared() {
		case DInt(1):
			return 1
		case DInt(2):
			switch td.Conns[2].DistanceSquared() {
			case DInt(3):
				return 2
			case DInt(5):
				return 3
			}
		}
	} else {
		switch td.Conns[1].DistanceSquared() {
		case DInt(2):
			return 0
		case DInt(3):
			switch td.Conns[2].DistanceSquared() {
			case DInt(3):
				return 4
			case DInt(5):
				return 5
			}
		case DInt(5):
			return 6
		}
	}
	Log.Errorf("Did not find correct index for %v", *td)
	return -1
}

/***************************************************************/
// PointPackData Functions for GetTrioDetails
/***************************************************************/

func (ppd *BasePointPackData) GetTrioDetails(trIdx TrioIndex) *TrioDetails {
	ppd.checkTrioInitialized()
	return ppd.AllTrioDetails[trIdx]
}


