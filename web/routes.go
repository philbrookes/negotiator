package web

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/feedhenry/negotiator/deploy"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// BuildRouter is the main place we build the mux router
func BuildRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	return r
}

// BuildHTTPHandler constructs a http.Handler it is also where common middleware is added via negroni
func BuildHTTPHandler(r *mux.Router) http.Handler {
	//recovery middleware for any panics in the handlers
	recovery := negroni.NewRecovery()
	recovery.PrintStack = false
	//add middleware for all routes
	n := negroni.New(recovery)
	n.UseFunc(CorrellationID)
	auth := Auth{logger: logrus.StandardLogger()}
	n.UseFunc(auth.Auth)
	// set up sys routes
	n.UseHandler(r)
	return n
}

// DeployRoute sets up the deploy route
func DeployRoute(r *mux.Router, logger Logger, controller *deploy.Controller, clientFactory DeployClientFactory) {
	deployHandler := NewDeployHandler(logger, controller, clientFactory)
	r.HandleFunc("/deploy/{nameSpace}/{template}", prometheus.InstrumentHandlerFunc("deploy", deployHandler.Deploy)).Methods("POST")
}

// SysRoute sets up the sys routes
func SysRoute(r *mux.Router) {
	sysHandler := SysHandler{}
	r.HandleFunc("/sys/info/ping", prometheus.InstrumentHandlerFunc("ping", sysHandler.Ping)).Methods("GET")
	r.HandleFunc("/sys/info/health", prometheus.InstrumentHandlerFunc("health", sysHandler.Health)).Methods("GET")
}

// Metrics route
func Metrics(r *mux.Router) {
	r.Handle("/metrics", prometheus.Handler()).Methods("GET")
}