package m3server

import (
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"testing"
)

func TestWriteAllTables(t *testing.T) {
	GenerateTextFilesEnv(pointdb.GetServerFullTestDb(m3util.PointTestEnv))
}
