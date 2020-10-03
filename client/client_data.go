package client

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
)

type ClientPointPackData struct {
	m3point.BasePointPackData
	env                 *QsmApiEnvironment
	ValidNextTrio       [12][2]m3point.TrioIndex
	AllMod4Permutations [12][4]m3point.TrioIndex
	AllMod8Permutations [12][8]m3point.TrioIndex
}

/***************************************************************/
// ClientConnection Functions
/***************************************************************/

func GetClientPointPackData(env m3util.QsmEnvironment) *ClientPointPackData {
	return env.GetData(m3util.PointIdx).(*ClientPointPackData)
}

func GetClientPathPackData(env m3util.QsmEnvironment) *ClientPathPackData {
	return env.GetData(m3util.PathIdx).(*ClientPathPackData)
}

func GetClientSpacePackData(env m3util.QsmEnvironment) *ClientSpacePackData {
	return env.GetData(m3util.SpaceIdx).(*ClientSpacePackData)
}

/***************************************************************/
// ClientPointPackData Functions for GetTrioDetails
/***************************************************************/

func (ppd *ClientPointPackData) GetValidNextTrio() [12][2]m3point.TrioIndex {
	return ppd.ValidNextTrio
}

func (ppd *ClientPointPackData) GetAllMod4Permutations() [12][4]m3point.TrioIndex {
	return ppd.AllMod4Permutations
}

func (ppd *ClientPointPackData) GetAllMod8Permutations() [12][8]m3point.TrioIndex {
	return ppd.AllMod8Permutations
}
