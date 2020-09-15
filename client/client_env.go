package client

import (
	"fmt"
	"github.com/freddy33/qsm-go/client/config"
	"github.com/freddy33/qsm-go/m3util"
	"io/ioutil"
	"net/http"
	"strings"
)

var Log = m3util.NewLogger("client", m3util.INFO)

type ClientConnection struct {
	BackendRootURL string
	EnvId          m3util.QsmEnvID
}

func newClientConnection(config config.Config, envId m3util.QsmEnvID) *ClientConnection {
	result := new(ClientConnection)
	result.BackendRootURL = config.BackendRootURL
	result.EnvId = envId
	result.validate()
	return result
}

func (cl *ClientConnection) validate() {
	if cl.EnvId < 1 {
		Log.Fatalf("Invalid client env id " + cl.EnvId.String() + " for root URL: " + cl.BackendRootURL)
	}
	if len(cl.BackendRootURL) < 4 {
		Log.Fatalf("Invalid client root URL: " + cl.BackendRootURL)
	}
	if !strings.HasSuffix(cl.BackendRootURL, "/") {
		cl.BackendRootURL = cl.BackendRootURL + "/"
	}
}

type QsmApiEnvironment struct {
	m3util.BaseQsmEnvironment
	clConn *ClientConnection
}

func (env *QsmApiEnvironment) InternalClose() error {
	Log.Infof("Closing API environment %d", env.GetId())
	return nil
}

func createNewApiEnv(envId m3util.QsmEnvID) m3util.QsmEnvironment {
	env := QsmApiEnvironment{}
	env.Id = envId

	clientConfig := config.NewConfig()
	env.clConn = newClientConnection(clientConfig, envId)

	return &env
}

func getEnvironment(envId m3util.QsmEnvID) *QsmApiEnvironment {
	return m3util.GetEnvironmentWithCreator(envId, createNewApiEnv).(*QsmApiEnvironment)
}

func GetInitializedApiEnv(envId m3util.QsmEnvID) *QsmApiEnvironment {
	env := getEnvironment(envId)
	cl := env.clConn

	if !cl.CheckServerUp() {
		Log.Fatalf("Test backend server down!")
	}

	if m3util.TestMode {
		// Equivalent of calling filldb job
		body := cl.ExecReq(http.MethodPost, "test-init", nil)
		defer m3util.CloseBody(body)
		b, err := ioutil.ReadAll(body)
		if err != nil {
			Log.Errorf("Could not read body from REST API end point %q due to %s", "test-init", err.Error())
			return nil
		}
		response := string(b)
		substr := fmt.Sprintf("env id %d was initialized", cl.EnvId)
		if strings.Contains(response, substr) {
			Log.Debugf("All good on home response %q", response)
		} else {
			Log.Errorf("The response from REST API end point %q did not have %s in %q", "test-init", substr, response)
			return nil
		}
	}

	env.initializePointData()
	env.initializePathData()
	return env
}
