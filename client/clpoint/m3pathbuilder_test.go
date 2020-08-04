package clpoint

import (
	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisplayPathBuilders(t *testing.T) {
	m3point.Log.SetAssert(true)
	m3util.SetToTestMode()

	env := client.GetFullApiTestEnv(m3util.PointTestEnv)
	ppd := client.GetApiPointPackData(env)
	assert.Equal(t, m3point.TotalNumberOfCubes+1, len(ppd.PathBuilders))
	growthCtx := ppd.GetGrowthContextByTypeAndIndex(m3point.GrowthType(8), 0)
	pnb := ppd.GetPathNodeBuilder(growthCtx, 0, m3point.Origin)
	assert.NotNil(t, pnb, "did not find builder for %s", growthCtx.String())
	rpnb, tok := pnb.(*m3point.RootPathNodeBuilder)
	assert.True(t, tok, "%s is not a root builder", pnb.String())
	m3point.Log.Debug(rpnb.DumpInfo())
}
