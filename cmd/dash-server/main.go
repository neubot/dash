// dash-server is the dash command line server.
//
// Usage:
//
//    dash-server [-datadir <datadir>]
//
// The server will listen on port 80 for incoming DASH requests
// to serve from DASH clients. The `datadir <datadir>` flag specifies
// the directory where to write measurement results.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	dash "github.com/neubot/dash/server"
)

var (
	flagDatadir = flag.String("datadir", ".", "directory where to save results")
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	handler := dash.NewHandler(*flagDatadir)
	handler.StartReaper(context.Background())
	handler.RegisterHandlers(mux)
	log.Fatal(http.ListenAndServe(":80", mux))
}
