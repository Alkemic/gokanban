package app

import (
	"log"
	"net/http"
)

type restHandler interface {
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
