package pathdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/stretchr/testify/assert"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPointsTable(t *testing.T) {
	m3util.SetToTestMode()
	env := GetPathDbCleanEnv(m3util.PathTempEnv)

	te, err := env.GetOrCreateTableExec(PointsTable)
	if !assert.NoError(t, err) {
		return
	}
	err = te.PrepareQueries()
	if !assert.NoError(t, err) {
		return
	}

	// Insert and select [1,2,3]
	pid, err := te.InsertReturnId(1, 2, 3)
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, pid > 0)
	rows, err := te.Query(FindPointIdPerCoord, 1, 2, 3)
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, rows.Next())
	var pid2 int64
	err = rows.Scan(&pid2)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, pid2, pid)
	assert.False(t, rows.Next())

	// Test unique point constraint
	pid3, err := te.InsertReturnId(1, 2, 3)
	assert.NotNil(t, err)
	errorMessage := err.Error()
	assert.True(t, strings.Contains(errorMessage, "duplicate key"))
	assert.True(t, strings.Contains(errorMessage, "points_x_y_z_key"))
	assert.Equal(t, int64(-1), pid3)

	// insert -1,2,3 and show next and new id from before
	pid4, err := te.InsertReturnId(-1, 2, 3)
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, pid4 > pid)

	// select -1,-2,-3 should return no rows
	rows2, err := te.Query(FindPointIdPerCoord, -1, -2, -3)
	if !assert.NoError(t, err) {
		return
	}
	assert.False(t, rows2.Next())
}

func TestPointsTableConcurrency(t *testing.T) {
	runtime.GOMAXPROCS(16)
	m3util.SetToTestMode()
	env := GetPathDbCleanEnv(m3util.PerfTestEnv)
	pathData := GetServerPathPackData(env)
	// increase concurrency chance with low random
	rdMax := m3point.CInt(100)
	nbRoutines := 50
	nbRound := 500
	start := time.Now()
	wg := new(sync.WaitGroup)
	for r := 0; r < nbRoutines; r++ {
		wg.Add(1)
		go func() {
			for i := 0; i < nbRound; i++ {
				randomPoint := m3point.CreateRandomPoint(rdMax)
				id := pathData.GetOrCreatePoint(randomPoint)
				assert.True(t, id > 0)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	Log.Infof("It took %v to create %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
}
