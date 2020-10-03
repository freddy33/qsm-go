package client

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/model/m3space"
)

type ClientSpacePackData struct {
	env        *QsmApiEnvironment
	allSpaces map[int]*SpaceCl
}

type SpaceCl struct {
	spaceData *ClientSpacePackData

	id   int
	name string

	maxCoord m3point.CInt
	maxTime  m3space.DistAndTime

	activePathNodeThreshold m3space.DistAndTime
	maxTriosPerPoint        int
	maxPathNodesPerPoint    int
}

/***************************************************************/
// ClientSpacePackData Functions
/***************************************************************/

func (spd *ClientSpacePackData) GetEnvId() m3util.QsmEnvID {
	panic("implement me")
}

func (spd *ClientSpacePackData) GetAllSpaces() []m3space.SpaceIfc {
	panic("implement me")
}

func (spd *ClientSpacePackData) GetSpace(id int) m3space.SpaceIfc {
	panic("implement me")
}

func (spd *ClientSpacePackData) CreateSpace(name string, activePathNodeThreshold m3space.DistAndTime, maxTriosPerPoint int, maxPathNodesPerPoint int) (m3space.SpaceIfc, error) {
	panic("implement me")
}

/***************************************************************/
// SpaceCl Functions
/***************************************************************/

func (s *SpaceCl) GetId() int {
	panic("implement me")
}

func (s *SpaceCl) GetName() string {
	panic("implement me")
}

func (s *SpaceCl) GetActivePathNodeThreshold() m3space.DistAndTime {
	panic("implement me")
}

func (s *SpaceCl) GetMaxTriosPerPoint() int {
	panic("implement me")
}

func (s *SpaceCl) GetMaxPathNodesPerPoint() int {
	panic("implement me")
}

func (s *SpaceCl) GetEvent(id m3space.EventId) m3space.EventIfc {
	panic("implement me")
}

func (s *SpaceCl) GetActiveEventsAt(time m3space.DistAndTime) []m3space.EventIfc {
	panic("implement me")
}

func (s *SpaceCl) GetSpaceTimeAt(time m3space.DistAndTime) m3space.SpaceTimeIfc {
	panic("implement me")
}

func (s *SpaceCl) CreateEvent(growthType m3point.GrowthType, growthIndex int, growthOffset int, creationTime m3space.DistAndTime, center m3point.Point, color m3space.EventColor) (m3space.EventIfc, error) {
	panic("implement me")
}

