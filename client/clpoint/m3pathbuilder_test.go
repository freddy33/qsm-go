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

	env := client.GetInitializedApiEnv(m3util.TestClientEnv)
	ppd := client.GetClientPointPackData(env)
	assert.Equal(t, m3point.TotalNbContexts, len(ppd.AllGrowthContexts))
	growthCtx := ppd.GetGrowthContextByTypeAndIndex(m3point.GrowthType(8), 0)
	assert.NotNil(t, growthCtx)
}
