package m3server

import (
	"context"
	"fmt"
	"github.com/freddy33/qsm-go/model/m3point"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const (
	QSM_CTX_ENV_ID_KEY = "QsmEnvId"
)

type QsmApp struct {
	Router *mux.Router
	Env    *m3db.QsmDbEnvironment
}

func (app *QsmApp) AddHandler(path string, handleFunc func(http.ResponseWriter, *http.Request)) {
	app.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), QSM_CTX_ENV_ID_KEY, app.Env.GetId())
		handleFunc(w, r.WithContext(ctx))
	})
}

func GetEnvironment(r *http.Request) *m3db.QsmDbEnvironment {
	return m3util.GetEnvironment(r.Context().Value(QSM_CTX_ENV_ID_KEY).(m3util.QsmEnvID)).(*m3db.QsmDbEnvironment)
}

func home(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "REST APIs at /point-data\nUsing env id=%d", r.Context().Value(QSM_CTX_ENV_ID_KEY))
	if err != nil {
		log.Printf("failed to send data to response due to %q", err.Error())
	}
}

func MakeApp(envId m3util.QsmEnvID) *QsmApp {
	var env *m3db.QsmDbEnvironment
	if envId == m3util.NoEnv {
		env = m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment)
		envId = env.GetId()
	} else {
		env = m3util.GetEnvironment(envId).(*m3db.QsmDbEnvironment)
	}
	m3point.InitializeDBEnv(env, false)

	r := mux.NewRouter()
	app := &QsmApp{Router: r, Env: env}
	app.AddHandler("/", home)
	app.AddHandler("/point-data", GetPointData)

	return app
}
