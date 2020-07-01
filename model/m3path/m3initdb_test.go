package m3path

import (
	"github.com/freddy33/qsm-go/utils/m3util"
	"testing"
)

func TestCreatePathTable(t *testing.T) {
	m3util.SetToTestMode()
	createTablesEnv(GetCleanTempDb(m3util.PathTempEnv))
}

