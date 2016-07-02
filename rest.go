package main

import "net/http"

type ViewFunction func(http.ResponseWriter, *http.Request, map[string]string)

type RESTEndPoint struct {
	Get    ViewFunction
	Post   ViewFunction
	Put    ViewFunction
	Delete ViewFunction

	Options ViewFunction
}

func (rest *RESTEndPoint) Dispatch(
	w http.ResponseWriter,
	r *http.Request,
	p map[string]string,
) {
	if r.Method == "GET" && rest.Get != nil {
		rest.Get(w, r, p)
	} else if r.Method == "POST" && rest.Post != nil {
		rest.Post(w, r, p)
	} else if r.Method == "PUT" && rest.Put != nil {
		rest.Put(w, r, p)
	} else if r.Method == "DELETE" && rest.Delete != nil {
		rest.Delete(w, r, p)
	} else if r.Method == "OPTIONS" && rest.Options != nil {
		rest.Options(w, r, p)
	}
}
