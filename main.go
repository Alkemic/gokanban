package main

import (
	"flag"
	"log"
	"os"
)

var (
	bindHost = flag.String("host", "", "")
	dbName   = flag.String("db-name", "./db.sqlite", "")
	bindPort = flag.Int("port", 8080, "")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile|log.Ldate)

	application := NewApp(logger, *bindHost, *bindPort, *dbName)
	application.Run()
}
