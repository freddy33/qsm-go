package m3path

import (
	"github.com/freddy33/qsm-go/m3db"
	"testing"
)

func TestCreatePathTable(t *testing.T) {
	m3db.SetToTestMode()
	createTablesEnv(GetCleanTempDb(m3db.PathTempEnv))
}

