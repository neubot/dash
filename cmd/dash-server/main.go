// dash-server is the dash command line server.
//
// Usage:
//
//    dash-server [-datadir <datadir>] [-endpoint <endpoint>]
//
// The server will listen for incoming DASH experiment requests and
// will keep serving them until it is interrupted.
//
// The `-datadir <datadir>` flag specifies the directory where to write
// measurement results. By default is the current working directory.
//
// The `-endpoint <endpoint>` flag specifies the endpoint to listen
// to for unencrypted DASH experiment requests. By default we will
// listen on `:80`, i.e., on port `80` on all interfaces.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/neubot/dash/server"
)

var (
	flagDatadir  = flag.String("datadir", ".", "directory where to save results")
	flagEndpoint = flag.String("endpoint", ":80", "endpoint where to listen")
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	handler := server.NewHandler(*flagDatadir)
	handler.StartReaper(context.Background())
	handler.RegisterHandlers(mux)
	log.Fatal(http.ListenAndServe(*flagEndpoint, mux))
}
