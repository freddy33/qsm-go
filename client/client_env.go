package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/client/config"
	"github.com/freddy33/qsm-go/m3util"
	"net/http"
	"strings"
	"time"
)

var Log = m3util.NewLogger("client", m3util.INFO)

type ClientConnection struct {
	backendRootURL string
	envId          m3util.QsmEnvID
	httpClient     http.Client
}

type QsmApiEnvironment struct {
	m3util.BaseQsmEnvironment
	clConn *ClientConnection
}

func createNewApiEnv(envId m3util.QsmEnvID) m3util.QsmEnvironment {
	env := QsmApiEnvironment{}
	env.Id = envId

	clientConfig := config.NewConfig()
	result := new(ClientConnection)
	result.backendRootURL = clientConfig.BackendRootURL
	result.envId = envId
	result.validate()
	env.clConn = result

	return &env
}

func (cl *ClientConnection) validate() {
	if cl.envId < 1 {
		Log.Fatalf("Invalid client env id " + cl.envId.String() + " for root URL: " + cl.backendRootURL)
	}
	if len(cl.backendRootURL) < 4 {
		Log.Fatalf("Invalid client root URL: " + cl.backendRootURL)
	}
	if !strings.HasSuffix(cl.backendRootURL, "/") {
		cl.backendRootURL = cl.backendRootURL + "/"
	}
	cl.httpClient = http.Client{Timeout: 20 * time.Second}
}

func (env *QsmApiEnvironment) Close() {
	Log.Infof("Closing API environment %d", env.GetId())
	env.clConn.httpClient.CloseIdleConnections()
}

func GetInitializedApiEnv(envId m3util.QsmEnvID) *QsmApiEnvironment {
	return GetOrCreateInitializedApiEnv(envId, false, m3util.TestMode)
}

func GetOrCreateInitializedApiEnv(envId m3util.QsmEnvID, callDrop, callInit bool) *QsmApiEnvironment {
	env := m3util.GetEnvironmentWithCreator(envId, createNewApiEnv).(*QsmApiEnvironment)
	cl := env.clConn

	if !cl.CheckServerUp() {
		Log.Fatalf("Test backend server down!")
		return nil
	}

	if callDrop {
		// Equivalent of calling filldb job
		uri := "drop-env"
		response, err := cl.ExecReq(http.MethodDelete, uri, nil, nil, false)
		if err != nil {
			Log.Fatal(err)
			return nil

		}
		substr := fmt.Sprintf("env id %d was deleted", cl.envId)
		if strings.Contains(response, substr) {
			Log.Debugf("All good on home response %q", response)
		} else {
			Log.Fatalf("The response from REST API end point %q did not have %s in %q", uri, substr, response)
			return nil
		}
		env.SetData(m3util.PointIdx, nil)
		env.SetData(m3util.PathIdx, nil)
		env.SetData(m3util.SpaceIdx, nil)
	}

	if callInit {
		// Equivalent of calling filldb job
		uri := "init-env"
		response, err := cl.ExecReq(http.MethodPost, uri, nil, nil, false)
		if err != nil {
			Log.Fatal(err)
			return nil
		}
		substr := fmt.Sprintf("env id %d was initialized", cl.envId)
		if strings.Contains(response, substr) {
			Log.Debugf("All good on home response %q", response)
		} else {
			Log.Fatalf("The response from REST API end point %q did not have %s in %q", uri, substr, response)
			return nil
		}
	}

	env.initializePointData()
	env.initializePathData()
	env.initializeSpaceData()
	return env
}


