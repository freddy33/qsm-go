package m3db

import (
	"database/sql"
	"fmt"
	"sync"

	config "github.com/freddy33/qsm-go/backend/conf"

	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
)

var Log = m3util.NewLogger("m3db", m3util.INFO)

type DbConnDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
}

type QsmError string

func MakeQsmErrorf(format string, args ...interface{}) QsmError {
	return QsmError(fmt.Sprintf(format, args...))
}

func (qsmError QsmError) Error() string {
	return string(qsmError)
}

type QsmDbEnvironment struct {
	m3util.BaseQsmEnvironment
	dbDetails        DbConnDetails
	db               *sql.DB
	createTableMutex sync.Mutex
	tableExecs       map[string]*TableExec
}

func (env *QsmDbEnvironment) GetConnection() *sql.DB {
	return env.db
}

func (env *QsmDbEnvironment) GetDbConf() DbConnDetails {
	return env.dbDetails
}

func createNewDbEnv(envId m3util.QsmEnvID) m3util.QsmEnvironment {
	env := QsmDbEnvironment{}
	env.Id = envId
	env.tableExecs = make(map[string]*TableExec)

	env.fillDbConf()
	env.openDb()

	if !env.Ping() {
		Log.Fatalf("Could not ping DB %d", envId)
	}

	return &env
}

func GetEnvironment(envId m3util.QsmEnvID) *QsmDbEnvironment {
	return m3util.GetEnvironmentWithCreator(envId, createNewDbEnv).(*QsmDbEnvironment)
}

func (env *QsmDbEnvironment) fillDbConf() {
	env.dbDetails = DbConnDetails{
		Host:     config.DbHost,
		Port:     config.DbPort,
		User:     config.DbUser,
		Password: config.DbPassword,
		DbName:   config.DbName,
	}

	if Log.IsDebug() {
		Log.Debugf("DB conf for environment %d is user=%s dbName=%s", env.GetId(), env.dbDetails.User, env.dbDetails.DbName)
	}
}

func (env *QsmDbEnvironment) openDb() {
	connDetails := env.GetDbConf()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connDetails.Host, connDetails.Port, connDetails.User, connDetails.Password, connDetails.DbName)
	if Log.IsDebug() {
		Log.Debugf("Opening DB for environment %d is user=%s dbName=%s", env.GetId(), env.dbDetails.User, env.dbDetails.DbName)
	}
	var err error
	env.db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		Log.Fatalf("fail to open DB for environment %d with user=%s and dbName=%s due to %v", env.GetId(), env.dbDetails.User, env.dbDetails.DbName, err)
	}
	if Log.IsDebug() {
		Log.Debugf("DB opened for environment %d is user=%s dbName=%s", env.GetId(), env.dbDetails.User, env.dbDetails.DbName)
	}
}

func (env *QsmDbEnvironment) InternalClose() error {
	envId := env.GetId()
	Log.Infof("Closing DB environment %d", envId)
	defer m3util.RemoveEnvFromMap(envId)
	env.CleanAllData()
	// clean table exec
	for tn, te := range env.tableExecs {
		err := te.Close()
		if err != nil {
			Log.Warnf("Closing table exec of envId=%d table=%s generated '%s'", env.GetId(), tn, err.Error())
		}
		delete(env.tableExecs, tn)
	}
	db := env.db
	env.db = nil
	if db != nil {
		return db.Close()
	}
	return nil
}

func (env *QsmDbEnvironment) Destroy() {
	err := env.InternalClose()
	if err != nil {
		Log.Error(err)
	}
	m3util.RunQsm(env.GetId(), "db", "drop")
}

func (env *QsmDbEnvironment) Ping() bool {
	err := env.GetConnection().Ping()
	if err != nil {
		Log.Errorf("failed to ping %d on DB %s due to %v", env.GetId(), env.dbDetails.DbName, err)
		return false
	}
	if Log.IsDebug() {
		Log.Debugf("ping for environment %d successful", env.GetId())
	}
	return true
}
