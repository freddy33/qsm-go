package m3server

import (
	"context"
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const (
	QSM_CTX_ENV_ID_KEY = "QsmEnvId"
)

type QsmApp struct {
	Router *mux.Router
	Env    *m3db.QsmEnvironment
}

func (app *QsmApp) AddHandler(path string, handleFunc func(http.ResponseWriter, *http.Request)) {
	app.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), QSM_CTX_ENV_ID_KEY, app.Env.GetId())
		handleFunc(w, r.WithContext(ctx))
	})
}

func GetEnvironment(r *http.Request) *m3db.QsmEnvironment {
	return m3db.GetEnvironment(r.Context().Value(QSM_CTX_ENV_ID_KEY).(m3db.QsmEnvID))
}

func home(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "REST APIs at /point-data\nUsing env id=%d", r.Context().Value(QSM_CTX_ENV_ID_KEY))
	if err != nil {
		log.Printf("failed to send data to response due to %q", err.Error())
	}
}

func MakeApp(envId m3db.QsmEnvID) *QsmApp {
	var env *m3db.QsmEnvironment
	if envId == m3db.NoEnv {
		env = m3db.GetDefaultEnvironment()
		envId = env.GetId()
	} else {
		env = m3db.GetEnvironment(envId)
	}
	m3point.InitializeDBEnv(env, false)

	r := mux.NewRouter()
	app := &QsmApp{Router: r, Env: env}
	app.AddHandler("/", home)
	app.AddHandler("/point-data", GetPointData)

	return app
}
