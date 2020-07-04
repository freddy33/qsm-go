package m3api

import "github.com/freddy33/qsm-go/utils/m3util"

var Log = m3util.NewLogger("m3api", m3util.INFO)

type QsmApiEnvironment struct {
	m3util.BaseQsmEnvironment
}

func (env *QsmApiEnvironment) InternalClose() error {
	Log.Infof("Closing API environment %d", env.GetId())
	return nil
}

func createNewEnv(envId m3util.QsmEnvID) m3util.QsmEnvironment {
	env := QsmApiEnvironment{}
	env.Id = envId

	return &env
}

func SetEnvironmentCreator() {
	m3util.SetEnvironmentCreator(createNewEnv)
}