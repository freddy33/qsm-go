package clpoint

import (
	"testing"

	"github.com/freddy33/qsm-go/client"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
)

var Log = m3util.NewLogger("clpoint", m3util.INFO)

func TestDisplayPathBuilders(t *testing.T) {
	Log.SetAssert(true)
	m3util.SetToTestMode()

	env := client.GetInitializedApiEnv(m3util.PointTestEnv)
	ppd := client.GetClientPointPackData(env)
	assert.Equal(t, m3point.TotalNumberOfCubes+1, len(ppd.PathBuilders))
	growthCtx := ppd.GetGrowthContextByTypeAndIndex(m3point.GrowthType(8), 0)
	pnb := ppd.GetPathNodeBuilder(growthCtx, 0, m3point.Origin)
	assert.NotNil(t, pnb, "did not find builder for %s", growthCtx.String())
	rpnb, tok := pnb.(*m3point.RootPathNodeBuilder)
	assert.True(t, tok, "%s is not a root builder", pnb.String())
	Log.Debug(rpnb.DumpInfo())
}
