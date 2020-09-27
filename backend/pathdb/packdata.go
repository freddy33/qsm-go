package pathdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3path"
	"github.com/freddy33/qsm-go/model/m3point"
	"sync"
)

type ServerPathPackData struct {
	env        *m3db.QsmDbEnvironment
	pathCtxMap map[int]*PathContextDb

	pointsTe    *m3db.TableExec
	pathCtxTe   *m3db.TableExec
	pathNodesTe *m3db.TableExec

	// All PathContexts centered at origin with growth type + offset
	AllCenterContexts       map[m3point.GrowthType][]*PathContextDb
	AllCenterContextsLoaded bool
}

func makeServerPathPackData(env m3util.QsmEnvironment) *ServerPathPackData {
	res := new(ServerPathPackData)
	res.env = env.(*m3db.QsmDbEnvironment)
	res.pathCtxMap = make(map[int]*PathContextDb, 2^8)
	res.AllCenterContexts = make(map[m3point.GrowthType][]*PathContextDb)
	res.AllCenterContextsLoaded = false
	return res
}

func GetServerPathPackData(env m3util.QsmEnvironment) *ServerPathPackData {
	if env.GetData(m3util.PathIdx) == nil {
		env.SetData(m3util.PathIdx, makeServerPathPackData(env))
	}
	return env.GetData(m3util.PathIdx).(*ServerPathPackData)
}

func (pathData *ServerPathPackData) GetEnvId() m3util.QsmEnvID {
	if pathData == nil {
		return m3util.NoEnv
	}
	return pathData.env.GetId()
}

func (pathData *ServerPathPackData) GetPathCtx(id int) m3path.PathContext {
	pathCtx, ok := pathData.pathCtxMap[id]
	if ok {
		return pathCtx
	}
	// TODO: Load from DB
	return nil
}

var allTestContextsMutex sync.Mutex

func (pathData *ServerPathPackData) GetAllPathContexts() map[m3point.GrowthType][]*PathContextDb {
	if pathData.AllCenterContextsLoaded {
		return pathData.AllCenterContexts
	}

	allTestContextsMutex.Lock()
	defer allTestContextsMutex.Unlock()

	if pathData.AllCenterContextsLoaded {
		return pathData.AllCenterContexts
	}

	pointData := pointdb.GetPointPackData(pathData.env)

	idx := 0
	for _, growthCtx := range pointData.GetAllGrowthContexts() {
		ctxType := growthCtx.GetGrowthType()
		maxOffset := ctxType.GetMaxOffset()
		if len(pathData.AllCenterContexts[ctxType]) == 0 {
			pathData.AllCenterContexts[ctxType] = make([]*PathContextDb, ctxType.GetNbIndexes()*maxOffset)
			idx = 0
		}
		for offset := 0; offset < maxOffset; offset++ {
			var err error
			pathData.AllCenterContexts[ctxType][idx], err = pathData.CreatePathCtxDb(growthCtx, offset)
			if err != nil {
				Log.Error(err)
				return nil
			}
			idx++
		}
	}

	pathData.AllCenterContextsLoaded = true
	return pathData.AllCenterContexts
}
