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
	allPathContextsLoadMutex sync.Mutex
	pathContextsLoaded       bool
	AllCenterContexts        map[m3point.GrowthType][]*PathContextDb
}

func makeServerPathPackData(env m3util.QsmEnvironment) *ServerPathPackData {
	res := new(ServerPathPackData)
	res.env = env.(*m3db.QsmDbEnvironment)
	res.pathCtxMap = make(map[int]*PathContextDb, 2^8)
	res.AllCenterContexts = make(map[m3point.GrowthType][]*PathContextDb)
	res.pathContextsLoaded = false
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

func (pathData *ServerPathPackData) GetPathCtxDb(id int) *PathContextDb {
	if !pathData.pathContextsLoaded {
		Log.Fatal("The path context should have been initialized with call to initAllPathContexts")
		return nil
	}
	pathCtx, ok := pathData.pathCtxMap[id]
	if ok {
		return pathCtx
	}
	// TODO: Load from DB
	return nil
}

func (pathData *ServerPathPackData) GetPathCtx(id int) m3path.PathContext {
	return pathData.GetPathCtxDb(id)
}

func (pathData *ServerPathPackData) addPathContext(pathCtx *PathContextDb) {
	growthType := pathCtx.GetGrowthType()
	nbIndexes := growthType.GetNbIndexes()
	maxOffset := growthType.GetMaxOffset()
	contexts := pathData.AllCenterContexts[growthType]
	if len(contexts) == 0 {
		allCtxSize := nbIndexes * maxOffset
		contexts = make([]*PathContextDb, allCtxSize)
		for i := 0; i < allCtxSize; i++ {
			contexts[i] = nil
		}
		pathData.AllCenterContexts[growthType] = contexts
	}
	allCtxIdx := pathCtx.GetGrowthIndex()*maxOffset + pathCtx.GetGrowthOffset()
	contexts[allCtxIdx] = pathCtx
}

func (pathData *ServerPathPackData) initAllPathContexts() error {
	if pathData.pathContextsLoaded {
		return nil
	}

	pathData.allPathContextsLoadMutex.Lock()
	defer pathData.allPathContextsLoadMutex.Unlock()

	if pathData.pathContextsLoaded {
		return nil
	}

	te := pathData.pathCtxTe
	nbPathCtx, toFill, err := te.GetForSaveAll()
	if err != nil {
		return err
	}
	if !toFill {
		Log.Info("Loading path contexts from DB")
		rows, err := te.SelectAllForLoad()
		if err != nil {
			return err
		}
		for rows.Next() {
			pathCtx, err := createPathCtxFromDbRows(rows, pathData)
			if err != nil {
				return err
			}
			pathData.addPathContext(pathCtx)
		}
	} else {
		Log.Info("Creating and saving all path contexts in DB")
		pointData := pointdb.GetPointPackData(pathData.env)
		for _, growthCtx := range pointData.GetAllGrowthContexts() {
			ctxType := growthCtx.GetGrowthType()
			maxOffset := ctxType.GetMaxOffset()
			for offset := 0; offset < maxOffset; offset++ {
				var err error
				pathCtx, err := pathData.internalCreatePathCtxDb(growthCtx, offset)
				if err != nil {
					return err
				}
				pathData.addPathContext(pathCtx)
			}
		}
		te.SetFilled()
	}

	nbPathCtx = len(pathData.pathCtxMap)
	Log.Infof("Environment %d has %d path contexts", pathData.env.GetId(), nbPathCtx)
	pathData.pathContextsLoaded = true
	return nil
}

func (pathData *ServerPathPackData) GetPathCtxFromAttributes(growthType m3point.GrowthType, growthIndex int, growthOffset int) (m3path.PathContext, error) {
	return pathData.GetPathCtxDbFromAttributes(growthType, growthIndex, growthOffset)
}

func (pathData *ServerPathPackData) GetPathCtxDbFromAttributes(growthType m3point.GrowthType, growthIndex int, growthOffset int) (*PathContextDb, error) {
	growthCtx := pointdb.GetPointPackData(pathData.env).GetGrowthContextByTypeAndIndex(growthType, growthIndex)
	if growthCtx == nil {
		return nil, m3util.MakeQsmErrorf("could not find Growth Context for %d %d", growthType, growthIndex)
	}
	err := pathData.initAllPathContexts()
	if err != nil {
		return nil, err
	}
	allCtxIdx := growthIndex*growthType.GetMaxOffset() + growthOffset
	contexts := pathData.AllCenterContexts[growthType]
	if len(contexts) > allCtxIdx {
		return contexts[allCtxIdx], nil
	}
	return nil, m3util.MakeQsmErrorf("could not find Path Context for %d %d %d", growthType, growthIndex, growthOffset)
}

func (pathData *ServerPathPackData) internalCreatePathCtxDb(growthCtx m3point.GrowthContext, offset int) (*PathContextDb, error) {
	pathCtx := PathContextDb{}
	pathCtx.pathData = pathData
	pathCtx.pointData = pointdb.GetPointPackData(pathData.env)
	pathCtx.growthCtx = growthCtx
	pathCtx.growthOffset = offset
	pathCtx.rootNode = nil
	pathCtx.maxDist = 0

	err := pathCtx.insertInDb()
	if err != nil {
		return nil, m3util.MakeWrapQsmErrorf(err, "could not save new path context %s due to %v", pathCtx.String(), err)
	}

	pathData.pathCtxMap[pathCtx.GetId()] = &pathCtx

	err = pathCtx.createRootNode()
	if err != nil {
		return nil, err
	}

	return &pathCtx, nil
}
