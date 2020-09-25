package m3server

import (
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"testing"
)

func TestWriteAllTables(t *testing.T) {
	m3util.SetToTestMode()
	GenerateTextFilesEnv(pointdb.GetPointDbFullEnv(m3util.PointTestEnv))
}
