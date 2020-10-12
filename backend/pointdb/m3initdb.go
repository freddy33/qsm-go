package pointdb

import (
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/m3util"
)

func GetPointDbFullEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	env := m3db.GetEnvironment(envId)

	err := env.ExecOnce(m3util.PointIdx, func() error {
		err := env.CheckSchema()
		if err != nil {
			return err
		}
		pointData := GetServerPointPackData(env)
		pointData.createTables()
		return nil
	})
	if err != nil {
		Log.Fatal(err)
		return nil
	}

	return env
}

// Do not use this environment to load
func GetPointDbCleanEnv(envId m3util.QsmEnvID) *m3db.QsmDbEnvironment {
	if !m3util.TestMode {
		Log.Fatalf("Cannot use GetPointDbCleanEnv in non test mode!")
	}

	env := m3db.GetEnvironment(envId)
	env.Destroy()

	return GetPointDbFullEnv(envId)
}
