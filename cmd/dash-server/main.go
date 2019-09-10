// dash-server is the dash command line server.
//
// Usage:
//
//    dash-server [-datadir <datadir>]
//                [-prometheusx.listen-address <endpoint>]
//                [-tls-key <key.pem>]
//                [-tls-cert <cert.pem>]
//
// The server will listen for incoming DASH experiment requests and
// will keep serving them until it is interrupted.
//
// It will listen on `:80` and `:443`. Note that TLS code assumes
// a `key.pem` and `cert.pem` files in the current directory. Adjust
// these files locations using `-tls-key` and `-tls-cert`.
//
// The `-datadir <datadir>` flag specifies the directory where to write
// measurement results. By default is the current working directory.
//
// The `-prometheusx.listen-address <endpoint>` flag controls the TCP
// endpoint where the server will expose Prometheus metrics.
//
// The server will emit access logs on the standard output using the
// usual format. The server will emit error logging on the standard
// error using github.com/apex/log's JSON format.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/gorilla/handlers"
	"github.com/m-lab/go/prometheusx"
	"github.com/m-lab/go/rtx"
	"github.com/neubot/dash/server"
)

var (
	flagDatadir  = flag.String("datadir", ".", "directory where to save results")
	flagTLSCert  = flag.String("tls-key", "cert.pem", "TLS certificate to use")
	flagTLSKey   = flag.String("tls-key", "key.pem", "TLS key to use")
)

func main() {
	log.Log = &log.Logger{
		Handler: json.New(os.Stderr),
		Level:   log.DebugLevel,
	}
	flag.Parse()
	promServer := prometheusx.MustServeMetrics()
	defer promServer.Close()
	mux := http.NewServeMux()
	handler := server.NewHandler(*flagDatadir)
	handler.StartReaper(context.Background())
	handler.RegisterHandlers(mux)
	handler.Logger = log.Log
	rootHandler := handlers.LoggingHandler(os.Stdout, mux)
	go func() {
		rtx.Must(http.ListenAndServeTLS(
			":443", *flagTLSCert, *flagTLSKey, rootHandler,
		), "Can't start HTTPS server")
	}()
	rtx.Must(http.ListenAndServe(":80", rootHandler), "Can't start HTTP server")
}
