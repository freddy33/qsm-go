package m3server

import (
	"github.com/freddy33/qsm-go/backend/spacedb"
	"github.com/freddy33/qsm-go/m3util"
	"testing"
)

func TestWriteAllTables(t *testing.T) {
	m3util.SetToTestMode()
	GenerateTextFilesEnv(spacedb.GetSpaceDbFullEnv(m3util.PointTestEnv))
}
