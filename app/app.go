package app

import (
	"log"
	"net/http"
)

type restHandler interface {
	TaskEndPointPost(rw http.ResponseWriter, req *http.Request, p map[string]string)
	TaskEndPointPut(rw http.ResponseWriter, req *http.Request, p map[string]string)
	TaskEndPointDelete(rw http.ResponseWriter, req *http.Request, p map[string]string)
	ColumnGet(rw http.ResponseWriter, req *http.Request, p map[string]string)
	ColumnList(rw http.ResponseWriter, req *http.Request, p map[string]string)
	GetMux() *http.ServeMux
}

type app struct {
	logger *log.Logger
	rest   restHandler
}

// NewApp returns new instance of app
func NewApp(logger *log.Logger, rest restHandler) *app {
	return &app{
		logger: logger,
		rest:   rest,
	}
}

func (a *app) Run(bindAddr string) {
	a.logger.Printf("Server starting on: %s\n", bindAddr)
	a.logger.Fatal(http.ListenAndServe(bindAddr, a.rest.GetMux()))
}
