package m3server

import (
	"github.com/freddy33/qsm-go/utils/m3util"
	"testing"
)

func TestWriteAllTables(t *testing.T) {
	GenerateTextFilesEnv(GetFullTestDb(m3util.PointTestEnv))
}