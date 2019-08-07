package m3path

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func getPointEnv(env *m3db.QsmEnvironment, pointId int64) *m3point.Point {
	te, err := env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Errorf("could not get points table exec due to %v", err)
		return nil
	}
	rows, err := te.Query(SelectPointPerId, pointId)
	if err != nil {
		Log.Errorf("could not select points table exec due to %v", err)
		return nil
	}
	defer te.CloseRows(rows)
	if rows.Next() {
		res := m3point.Point{}
		err = rows.Scan(&res[0], &res[1], &res[2])
		if err != nil {
			Log.Errorf("Could not read row of %s due to %v", PointsTable, err)
		} else {
			return &res
		}
	}
	return nil
}

func getOrCreatePointEnv(env *m3db.QsmEnvironment, p m3point.Point) int64 {
	te, err := env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Errorf("could not get points table exec due to %v", err)
		return -1
	}
	return getOrCreatePointTe(te, p)
}

func getOrCreatePointTe(te *m3db.TableExec, p m3point.Point) int64 {
	rows, err := te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
	if err != nil {
		Log.Errorf("could not select points table exec due to %v", err)
		return -1
	}
	defer te.CloseRows(rows)
	var id int64
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			Log.Errorf("could not convert points table id for %v due to %v", p, err)
			return -1
		}
		return id
	} else {
		id, err = te.InsertReturnId(p.X(), p.Y(), p.Z())
		if err == nil {
			return id
		} else {
			errorMessage := err.Error()
			if strings.Contains(errorMessage, "duplicate key") && strings.Contains(errorMessage, "points_x_y_z_key") {
				// got concurrent insert, let's just reselect
				rows, err = te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
				if err != nil {
					Log.Errorf("could not select points table for %v after duplicate key insert exec due to %v", p, err)
					return -1
				}
				defer te.CloseRows(rows)
				if !rows.Next() {
					Log.Errorf("selecting points table for %v after duplicate key returns no rows!", p)
				}
				err = rows.Scan(&id)
				if err != nil {
					Log.Errorf("could not convert points table id for %v due to %v", p, err)
					return -1
				}
				return id
			} else {
				Log.Errorf("got unknown points table for %v error %v", p, err)
				return -1
			}
		}
	}
}

/***************************************************************/
// perf test main
/***************************************************************/
func RunInsertRandomPoints() {
	m3db.SetToTestMode()
	env := GetFullTestDb(m3db.PerfTestEnv)
	// increase concurrency chance with low random
	rdMax := m3point.CInt(10)
	nbRoutines := 100
	nbRound := 250
	start := time.Now()
	wg := new(sync.WaitGroup)
	for r := 0; r < nbRoutines; r++ {
		wg.Add(1)
		go func() {
			for i := 0; i < nbRound; i++ {
				randomPoint := RandomPoint(rdMax)
				id := getOrCreatePointEnv(env, randomPoint)
				if id <= 0 {
					Log.Errorf("failed to insert %v got %d id", randomPoint, id)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	Log.Infof("It took %v to create %d points with nb routines=%d max coord %d", time.Now().Sub(start), nbRoutines*nbRound, nbRoutines, rdMax)
}

/***************************************************************/
// Utility methods for test
/***************************************************************/

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
