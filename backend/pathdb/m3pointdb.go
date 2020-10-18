package pathdb

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"strings"
)

func (pathData *ServerPathPackData) addToMap(pointId m3path.PointId, point m3point.Point) *m3path.PathPoint {
	return pathData.pathPointsMap.AddToMap(pointId, point)
}

func (pathData *ServerPathPackData) GetPoint(pointId m3path.PointId) (*m3path.PathPoint, error) {
	pp, ok := pathData.pathPointsMap.GetById(pointId)
	if ok {
		return pp, nil
	}
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
			return pathData.addToMap(pointId, res), nil
		}
	}
	return nil, m3util.MakeQsmErrorf("point id %d does not exists!", pointId)
}

func (pathData *ServerPathPackData) GetOrCreatePoint(p m3point.Point) (*m3path.PathPoint, error) {
	pp, ok := pathData.pathPointsMap.GetByPoint(p)
	if ok {
		return pp, nil
	}
	te := pathData.pointsTe
	rows, err := te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
	if err != nil {
		return nil, m3util.MakeWrapQsmErrorf(err, "could not select point %v in points table exec due to %v", p, err)
	}
	defer te.CloseRows(rows)
	var id int64
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, m3util.MakeWrapQsmErrorf(err, "could not read points table id for %v due to %v", p, err)
		}
		return pathData.addToMap(m3path.PointId(id), p), nil
	} else {
		id, err = te.InsertReturnId(p.X(), p.Y(), p.Z())
		if err == nil {
			return pathData.addToMap(m3path.PointId(id), p), nil
		} else {
			errorMessage := err.Error()
			if strings.Contains(errorMessage, "duplicate key") && strings.Contains(errorMessage, "points_x_y_z_key") {
				// got concurrent insert, let's just reselect
				rows, err = te.Query(FindPointIdPerCoord, p.X(), p.Y(), p.Z())
				if err != nil {
					return nil, m3util.MakeWrapQsmErrorf(err, "could not select points table for %v after duplicate key insert exec due to %v", p, err)
				}
				defer te.CloseRows(rows)
				if !rows.Next() {
					return nil, m3util.MakeQsmErrorf("selecting points table for %v after duplicate key returns no rows!", p)
				}
				err = rows.Scan(&id)
				if err != nil {
					return nil, m3util.MakeWrapQsmErrorf(err, "could not convert points table id for %v due to %v", p, err)
				}
				return pathData.addToMap(m3path.PointId(id), p), nil
			} else {
				return nil, m3util.MakeWrapQsmErrorf(err, "got unknown points table for %v error %v", p, err)
			}
		}
	}
}
