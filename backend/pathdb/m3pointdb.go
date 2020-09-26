package pathdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3point"
	"strings"
	"sync"
	"time"
)

func (pathData *ServerPathPackData) GetPoint(pointId int64) (*m3point.Point, error) {
	te := pathData.pointsTe
	rows, err := te.Query(SelectPointPerId, pointId)
	if err != nil {
		return nil, m3util.MakeWrapQsmErrorf(err, "could not select point %d from points table exec due to %v", pointId, err)
	}
	defer te.CloseRows(rows)
	if rows.Next() {
		res := m3point.Point{}
		err = rows.Scan(&res[0], &res[1], &res[2])
		if err != nil {
			return nil, m3util.MakeWrapQsmErrorf(err, "could not read row of %s for %d due to %v", PointsTable, pointId, err)
		} else {
			return &res, nil
		}
	}
	return nil, m3util.MakeQsmErrorf("point id %d does not exists!", pointId)
}

func (pathData *ServerPathPackData) GetOrCreatePoint(p m3point.Point) int64 {
	return getOrCreatePointTe(pathData.pointsTe, p)
}

func getOrCreatePointTe(te *m3db.TableExec, p m3point.Point) int64 {
	rows, err := te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
	if err != nil {
		Log.Fatalf("could not select points table exec due to %v", err)
		return -1
	}
	defer te.CloseRows(rows)
	var id int64
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			Log.Fatalf("could not convert points table id for %v due to %v", p, err)
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
					Log.Fatalf("could not select points table for %v after duplicate key insert exec due to %v", p, err)
					return -1
				}
				defer te.CloseRows(rows)
				if !rows.Next() {
					Log.Errorf("selecting points table for %v after duplicate key returns no rows!", p)
					return -1
				}
				err = rows.Scan(&id)
				if err != nil {
					Log.Errorf("could not convert points table id for %v due to %v", p, err)
					return -1
				}
				return id
			} else {
				Log.Fatalf("got unknown points table for %v error %v", p, err)
				return -1
			}
		}
	}
}

/***************************************************************/
// perf test main
/***************************************************************/
func RunInsertRandomPoints() {
	m3util.SetToTestMode()
	env := GetPathDbFullEnv(m3util.PerfTestEnv)
	pathData := GetServerPathPackData(env)
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
				randomPoint := m3point.CreateRandomPoint(rdMax)
				id := pathData.GetOrCreatePoint(randomPoint)
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
