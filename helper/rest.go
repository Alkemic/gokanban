package helper

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type ViewFunction func(http.ResponseWriter, *http.Request, map[string]string)

type RESTEndPoint struct {
	Get    ViewFunction
	Post   ViewFunction
	Put    ViewFunction
	Delete ViewFunction

	Options ViewFunction
}

func (rest *RESTEndPoint) Dispatch(w http.ResponseWriter, r *http.Request, p map[string]string) {
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

func TimeTrack(logger *log.Logger) func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func(start time.Time, name string, w http.ResponseWriter) {
				logger.Printf("%s took %s", name, time.Since(start))
			}(time.Now(), fmt.Sprintf("%s %s", r.Method, r.RequestURI), w)

			f(w, r)
		}
	}
}

func Handle500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
