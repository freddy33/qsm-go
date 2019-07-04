package m3path

import (
	"github.com/freddy33/qsm-go/m3db"
	"testing"
)

func TestPathDb(t *testing.T) {
	m3db.Log.SetTrace()
	Log.SetTrace()
	createTables()
}