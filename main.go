package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var bindHost, bindAddress string
var bindPort int

func init() {
	flag.StringVar(&bindHost, "host", "", "")
	flag.IntVar(&bindPort, "port", 8080, "")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	serveStatic := http.FileServer(http.Dir("."))
	http.Handle("/frontend/", serveStatic)

	TaskRouting := RegexpHandler{}
	TaskRouting.HandleFunc(`^/task/$`, TaskListView)
	TaskRouting.HandleFunc(`^/task/(?P<id>\d+)/$`, TaskView)
	http.HandleFunc("/task/", TimeTrackDecorator(TaskRouting.ServeHTTP))

	ColumnRouting := RegexpHandler{}
	ColumnRouting.HandleFunc(`^/column/$`, ColumnListView)
	ColumnRouting.HandleFunc(`^/column/(?P<id>\d+)/$`, ColumnView)
	http.HandleFunc("/column/", TimeTrackDecorator(ColumnRouting.ServeHTTP))

	http.HandleFunc("/",
		TimeTrackDecorator(func(w http.ResponseWriter, r *http.Request) {
			index, _ := ioutil.ReadFile("./frontend/templates/index.html")
			io.WriteString(w, string(index))
		}))

	bindAddress = fmt.Sprintf("%s:%d", bindHost, bindPort)
	log.Printf("Server starting on: %s\n", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, nil))
}
