package helper

import (
	"net/http"
)

type ViewFunction func(http.ResponseWriter, *http.Request)

type RESTEndPoint struct {
	Get    ViewFunction
	Post   ViewFunction
	Put    ViewFunction
	Delete ViewFunction

	Options ViewFunction
}

func (rest *RESTEndPoint) Dispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && rest.Get != nil {
		rest.Get(w, r)
	} else if r.Method == "POST" && rest.Post != nil {
		rest.Post(w, r)
	} else if r.Method == "PUT" && rest.Put != nil {
		rest.Put(w, r)
	} else if r.Method == "DELETE" && rest.Delete != nil {
		rest.Delete(w, r)
	} else if r.Method == "OPTIONS" && rest.Options != nil {
		rest.Options(w, r)
	}
}
