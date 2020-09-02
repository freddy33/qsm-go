package pathdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"testing"
)

func TestCreatePathTable(t *testing.T) {
	m3util.SetToTestMode()
	createTablesEnv(GetCleanTempDb(m3util.PathTempEnv))
}
