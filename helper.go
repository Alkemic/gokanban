package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func TimeTrackDecorator(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer TimeTrack(
			time.Now(),
			fmt.Sprintf(
				"%s %s",
				r.Method,
				r.RequestURI,
			),
		)
		f(w, r)
	}
}
