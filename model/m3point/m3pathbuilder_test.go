package m3point

import (
	"github.com/freddy33/qsm-go/utils/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisplayPathBuilders(t *testing.T) {
	Log.SetAssert(true)
	m3util.SetToTestMode()

	env := getApiFullTestEnv(m3util.PointTestEnv)
	ppd := getApiPointPackData(env)
	assert.Equal(t, TotalNumberOfCubes+1, len(ppd.PathBuilders))
	growthCtx := ppd.GetGrowthContextByTypeAndIndex(GrowthType(8), 0)
	pnb := ppd.GetPathNodeBuilder(growthCtx, 0, Origin)
	assert.NotNil(t, pnb, "did not find builder for %s", growthCtx.String())
	rpnb, tok := pnb.(*RootPathNodeBuilder)
	assert.True(t, tok, "%s is not a root builder", pnb.String())
	Log.Debug(rpnb.dumpInfo())
}

