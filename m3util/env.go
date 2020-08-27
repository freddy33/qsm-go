package m3util

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

type QsmEnvID int

const (
	NoEnv        QsmEnvID = iota // 0
	MainEnv                      // 1
	RunEnv                       // 2
	PerfTestEnv                  // 3
	ShellEnv                     // 4
	PointTestEnv                 // 5
	PathTestEnv                  // 6
	SpaceTestEnv                 // 7
	GlTestEnv                    // 8
	DbTempEnv                    // 9
	PointTempEnv                 // 10
	PathTempEnv                  // 11
	PointLoadEnv                 // 12
)

const (
	MaxNumberOfEnvironments = 15
	QsmEnvNumberKey         = "QSM_ENV_NUMBER"
)

const (
	PointIdx     = 0
	PathIdx      = 1
	SpaceIdx     = 2
	GlIdx        = 3
	MaxDataEntry = 4
)

type QsmDataPack interface {
	GetEnvId() QsmEnvID
}

type QsmEnvironment interface {
	GetId() QsmEnvID
	GetData(dataIdx int) QsmDataPack
	// TODO: This should move tho the env creator
	SetData(dataIdx int, dataPack QsmDataPack)
	InternalClose() error
}

type BaseQsmEnvironment struct {
	Id   QsmEnvID
	data [MaxDataEntry]QsmDataPack
}

func (envId QsmEnvID) String() string {
	return strconv.Itoa(int(envId))
}

func (env *BaseQsmEnvironment) GetId() QsmEnvID {
	return env.Id
}

func (env *BaseQsmEnvironment) GetEnvNumber() string {
	return env.Id.String()
}

func (env *BaseQsmEnvironment) GetData(dataIdx int) QsmDataPack {
	return env.data[dataIdx]
}

func (env *BaseQsmEnvironment) SetData(dataIdx int, dataPack QsmDataPack) {
	env.data[dataIdx] = dataPack
}

func (env *BaseQsmEnvironment) CleanAllData() {
	// clean data
	for i := 0; i < len(env.data); i++ {
		env.data[i] = nil
	}
}

var createEnvMutex sync.Mutex
var environments map[QsmEnvID]QsmEnvironment

var TestMode bool

func init() {
	environments = make(map[QsmEnvID]QsmEnvironment)
}

func SetToTestMode() {
	TestMode = true
}

func ReadEnvId(sourceInfo string, envIdStr string) QsmEnvID {
	id, err := strconv.ParseInt(envIdStr, 10, 16)
	if err != nil {
		Log.Fatalf("The variable from %s is not a DB number but %s", sourceInfo, envIdStr)
	}
	return QsmEnvID(id)
}

func GetDefaultEnvId() QsmEnvID {
	envId := MainEnv
	envIdFromOs := os.Getenv(QsmEnvNumberKey)
	if envIdFromOs != "" {
		envId = ReadEnvId(fmt.Sprintf("env var %q", QsmEnvNumberKey), envIdFromOs)
	}
	Log.Infof("Using default environment %d", envId)
	return envId
}

func SetDefaultEnvId(envId QsmEnvID) {
	ExitOnError(os.Setenv(QsmEnvNumberKey, envId.String()))
}

func GetEnvironmentWithCreator(envId QsmEnvID, createEnvFunc func(envId QsmEnvID) QsmEnvironment) QsmEnvironment {
	env, ok := environments[envId]
	if !ok {
		createEnvMutex.Lock()
		defer createEnvMutex.Unlock()
		env, ok = environments[envId]
		if !ok {
			env = createEnvFunc(envId)
			environments[envId] = env
		}
	}
	return env
}

func RemoveEnvFromMap(envId QsmEnvID) {
	createEnvMutex.Lock()
	defer createEnvMutex.Unlock()
	delete(environments, envId)
}

func CloseAll() {
	Log.Infof("Closing %d environment(s)", len(environments))
	toClose := make([]QsmEnvironment, 0, len(environments))
	for _, env := range environments {
		if env != nil {
			toClose = append(toClose, env)
		}
	}
	for _, env := range toClose {
		id := env.GetId()
		err := env.InternalClose()
		if err != nil {
			Log.Errorf("Error while closing environment %d", id)
		}
	}
}

func RunQsm(id QsmEnvID, params ...string) {
	osQsmCmd(id, params...)
}

func StartQsmBackend(id QsmEnvID, params ...string) *os.Process {
	return osStartBackend(id, params...)
}

func GetCompulsoryEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		message := fmt.Sprintf("missing %s", key)
		panic(message)
	}

	return value
}

func GetCompulsoryEnvInt(key string) int {
	valueString := GetCompulsoryEnv(key)
	valueInt, err := strconv.Atoi(valueString)
	if err != nil {
		panic("error parsing %s to int")
	}

	return valueInt
}
