package m3server

import (
	"github.com/freddy33/qsm-go/m3util"
	"testing"
)

func TestWriteAllTables(t *testing.T) {
	GenerateTextFilesEnv(getServerFullTestDb(m3util.PointTestEnv))
}