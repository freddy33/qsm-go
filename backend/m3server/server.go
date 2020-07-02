package m3server

import (
	"context"
	"fmt"
	"github.com/freddy33/qsm-go/utils/m3db"
	"github.com/freddy33/qsm-go/utils/m3util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const (
	HttpEnvIdKey = "QsmEnvId"
)

type QsmApp struct {
	Router *mux.Router
	Env    *m3db.QsmDbEnvironment
}

func (app *QsmApp) AddHandler(path string, handleFunc func(http.ResponseWriter, *http.Request)) {
	app.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		envId := app.Env.GetId()
		fromHeader := r.Header.Get(HttpEnvIdKey)
		if fromHeader == "" {
			r.Header.Add(HttpEnvIdKey, app.Env.GetEnvNumber())
		} else {
			envId = m3util.ReadEnvId(fmt.Sprintf("header var %q", HttpEnvIdKey), fromHeader)
		}
		ctx := context.WithValue(r.Context(), HttpEnvIdKey, envId)
		handleFunc(w, r.WithContext(ctx))
	})
}

func GetEnvId(r *http.Request) m3util.QsmEnvID {
	return r.Context().Value(HttpEnvIdKey).(m3util.QsmEnvID)
}

func GetEnvironment(r *http.Request) *m3db.QsmDbEnvironment {
	return m3util.GetEnvironment(GetEnvId(r)).(*m3db.QsmDbEnvironment)
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
	SendResponse(w, http.StatusOK, "REST APIs at /point-data\nUsing env id=%d", r.Context().Value(HttpEnvIdKey))
}

func drop(w http.ResponseWriter, r *http.Request) {
	env := GetEnvironment(r)
	envId := env.GetId()
	env.Destroy()
	SendResponse(w, http.StatusOK, "Test env id %d was deleted", envId)
}

func initialize(w http.ResponseWriter, r *http.Request) {
	envId := GetEnvId(r)
	env := GetFullTestDb(envId)
	InitializePointDBEnv(env, true)
	SendResponse(w, http.StatusCreated, "Test env id %d was initialized", envId)
}

func MakeApp(envId m3util.QsmEnvID) *QsmApp {
	var env *m3db.QsmDbEnvironment
	if envId == m3util.NoEnv {
		env = m3util.GetDefaultEnvironment().(*m3db.QsmDbEnvironment)
		envId = env.GetId()
	} else {
		env = m3util.GetEnvironment(envId).(*m3db.QsmDbEnvironment)
	}
	InitializePointDBEnv(env, false)

	r := mux.NewRouter()
	app := &QsmApp{Router: r, Env: env}
	app.AddHandler("/", home)
	app.AddHandler("/point-data", retrievePointData)
	app.AddHandler("/test-init", initialize)
	app.AddHandler("/test-drop", drop)

	return app
}
