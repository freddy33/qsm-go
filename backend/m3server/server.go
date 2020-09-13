package m3server

import (
	"context"
	"fmt"
	"github.com/freddy33/qsm-go/backend/m3db"
	"github.com/freddy33/qsm-go/backend/pointdb"
	"github.com/freddy33/qsm-go/m3util"
	"github.com/freddy33/qsm-go/model/m3api"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
)

type QsmApp struct {
	HttpServerDone *sync.WaitGroup
	Server         *http.Server
	Router         *mux.Router
	Env            *m3db.QsmDbEnvironment
}

func (app *QsmApp) AddHandler(path string, handleFunc func(http.ResponseWriter, *http.Request)) *mux.Route {
	return app.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		envId := app.Env.GetId()
		fromHeader := r.Header.Get(m3api.HttpEnvIdKey)
		if fromHeader == "" {
			r.Header.Add(m3api.HttpEnvIdKey, app.Env.GetEnvNumber())
		} else {
			envId = m3util.ReadEnvId(fmt.Sprintf("header var %q", m3api.HttpEnvIdKey), fromHeader)
		}
		ctx := context.WithValue(r.Context(), m3api.HttpEnvIdKey, envId)
		handleFunc(w, r.WithContext(ctx))
	})
}

func GetEnvId(r *http.Request) m3util.QsmEnvID {
	return r.Context().Value(m3api.HttpEnvIdKey).(m3util.QsmEnvID)
}

func GetEnvironment(r *http.Request) *m3db.QsmDbEnvironment {
	return m3db.GetEnvironment(GetEnvId(r))
}

func SendResponse(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		log.Printf("failed to send data to response due to %q", err.Error())
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	SendResponse(w, http.StatusOK, "Using env id=%d\nMethod=%s\n", r.Context().Value(m3api.HttpEnvIdKey), r.Method)
}

func drop(w http.ResponseWriter, r *http.Request) {
	env := GetEnvironment(r)
	envId := env.GetId()
	env.Destroy()
	SendResponse(w, http.StatusOK, "Test env id %d was deleted", envId)
}

func initialize(w http.ResponseWriter, r *http.Request) {
	envId := GetEnvId(r)
	env := pointdb.GetServerFullTestDb(envId)
	pointdb.InitializePointDBEnv(env, true)
	SendResponse(w, http.StatusCreated, "Test env id %d was initialized", envId)
}

func MakeApp(envId m3util.QsmEnvID) *QsmApp {
	if envId == m3util.NoEnv {
		envId = m3util.GetDefaultEnvId()
	}
	env := m3db.GetEnvironment(envId)
	pointdb.InitializePointDBEnv(env, false)

	r := mux.NewRouter()
	app := &QsmApp{Router: r, Env: env}
	app.AddHandler("/", home)
	app.AddHandler("/point-data", retrievePointData).Methods("GET")
	app.AddHandler("/test-init", initialize).Methods("POST")
	app.AddHandler("/test-drop", drop).Methods("DELETE")
	app.AddHandler("/create-path-ctx", createPathContext).Methods("PUT")
	app.AddHandler("/init-root-node", initRootNode).Methods("PUT")
	app.AddHandler("/next-nodes", moveToNextNode).Methods("POST")

	return app
}
