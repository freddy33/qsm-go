package m3path

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strings"
	"sync"
	"testing"
)

var cleanedDbMutex sync.Mutex
var cleanedDb bool
var tempEnv *m3db.QsmEnvironment

func getCleanTempDb() *m3db.QsmEnvironment {
	cleanedDbMutex.Lock()
	defer cleanedDbMutex.Unlock()

	if cleanedDb && tempEnv != nil && tempEnv.GetConnection() != nil {
		return tempEnv
	}
	tempEnv = m3db.GetEnvironment(m3db.TempEnv)
	tempEnv.Destroy()
	tempEnv = m3db.GetEnvironment(m3db.TempEnv)

	cleanedDb = true

	return tempEnv
}

func TestPathDb(t *testing.T) {
	createTables()
}

func TestPointsTable(t *testing.T) {
	env := getCleanTempDb()
	te, err := env.GetOrCreateTableExec(PointsTable)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// Insert and select [1,2,3]
	pid, err := te.InsertReturnId(1,2,3)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.True(t, pid > 0)
	rows, err := te.Query(FindPointIdPerCoord, 1,2,3)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.True(t, rows.Next())
	var pid2 int64
	err = rows.Scan(&pid2)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, pid2, pid)
	assert.False(t, rows.Next())

	// Test unique point constraint
	pid3, err := te.InsertReturnId(1,2,3)
	assert.NotNil(t, err)
	errorMessage := err.Error()
	assert.True(t, strings.Contains(errorMessage, "duplicate key"))
	assert.True(t, strings.Contains(errorMessage, "points_x_y_z_key"))
	assert.Equal(t, int64(-1), pid3)

	// insert -1,2,3 and show next and new id from before
	pid4, err := te.InsertReturnId(-1,2,3)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.True(t, pid4 > pid)

	// select -1,-2,-3 should return no rows
	rows2, err := te.Query(FindPointIdPerCoord, -1,-2,-3)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.False(t, rows2.Next())
}

func RandomPoint(max m3point.CInt) m3point.Point {
	return m3point.Point{RandomCInt(max), RandomCInt(max), RandomCInt(max)}
}

func RandomCInt(max m3point.CInt) m3point.CInt {
	r := m3point.CInt(rand.Int31n(int32(max)))
	if rand.Float32() < 0.5 {
		return -r
	}
	return r
}

func TestPointsTableConcurrency(t *testing.T) {
	// increase concurrency chance with low random
	rdMax := m3point.CInt(10)
	nbRoutines := 100
	nbRound := 20
	wg := new(sync.WaitGroup)
	for r:=0;r<nbRoutines;r++ {
		wg.Add(1)
		go func() {
			for i := 0; i <nbRound;i++ {
				randomPoint := RandomPoint(rdMax)
				id := GetOrCreatePoint(randomPoint)
				assert.True(t, id > 0)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
