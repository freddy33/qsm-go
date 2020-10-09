package pathdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
)

type OpenNodeBuilder struct {
	pathCtx                        *PathContextDb
	d                              int
	expectedSize                   int
	openNodesMap                   ServerPathNodeMap
	selectConflict, insertConflict int
}

func createCurrentNodeBuilder(pathCtx *PathContextDb) (*OpenNodeBuilder, error) {
	lastNodes, err := pathCtx.GetPathNodesAt(pathCtx.GetMaxDist())
	if err != nil {
		return nil, err
	}

	if len(lastNodes) == 0 {
		return nil, m3util.MakeQsmErrorf("cannot create open nodes builder from nothing at %s", pathCtx.String())
	}

	res := new(OpenNodeBuilder)
	res.pathCtx = pathCtx
	res.d = lastNodes[0].D()
	res.expectedSize = m3path.CalculatePredictedSize(pathCtx.GetGrowthType(), res.d)
	if res.expectedSize > 32 {
		res.openNodesMap = MakeHashPathNodeMap(res.expectedSize)
	} else {
		res.openNodesMap = MakeSimplePathNodeMap(res.expectedSize)
	}

	for i := 0; i < len(lastNodes); i++ {
		res.addPathNode(lastNodes[i].(*PathNodeDb))
	}

	return res, nil
}

func createNextNodeBuilder(previous *OpenNodeBuilder) *OpenNodeBuilder {
	res := new(OpenNodeBuilder)
	res.pathCtx = previous.pathCtx
	res.d = previous.d + 1
	res.expectedSize = m3path.CalculatePredictedSize(res.pathCtx.GetGrowthType(), res.d)
	if res.expectedSize > 32 {
		res.openNodesMap = MakeHashPathNodeMap(res.expectedSize)
	} else {
		res.openNodesMap = MakeSimplePathNodeMap(res.expectedSize)
	}
	return res
}

func (onb *OpenNodeBuilder) addPathNode(pn *PathNodeDb) *PathNodeDb {
	res, _ := onb.openNodesMap.AddPathNode(pn)
	return res
}

func (onb *OpenNodeBuilder) openNodesSize() int {
	return onb.openNodesMap.Size()
}
