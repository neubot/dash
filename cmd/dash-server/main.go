// dash-server is the dash command line server.
//
// Usage:
//
//    dash-server [-datadir <dirpath>]
//                [-http-listen-address <endpoint>]
//                [-https-listen-address <endpoint>]
//                [-prometheusx.listen-address <endpoint>]
//                [-tls-cert <filepath>]
//                [-tls-key <filepath>]
//
// The server will listen for incoming DASH experiment requests and
// will keep serving them until it is interrupted.
//
// By default the server listens for HTTP connections at `:8080` and
// for HTTPS connections at `:8443`. It assumes the TLS certificate
// is at `./cert.pem` and the TLS key is at `./key.pem`.
//
// The `-datadir <dirpath>` flag specifies the directory where to write
// measurement results. By default is the current working directory.
//
// The `-http-listen-address <endpoint>` flag allows to set the TCP endpoint
// where the server should listen for HTTP clients.
//
// The `-https-listen-address <endpoint>` flag allows to set the TCP endpoint
// where the server should listen for HTTPS clients.
//
// The `-prometheusx.listen-address <endpoint>` flag controls the TCP
// endpoint where the server will expose Prometheus metrics.
//
// The `-tls-cert <filepath>` flag allows to set the TLS certificate path.
//
// The `-tls-key <filepath>` flag allows to set the TLS key path.
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
	flagDatadir = flag.String(
		"datadir", ".", "directory where to save results",
	)
	flagHTTPListenAddress = flag.String(
		"http-listen-address", ":8080", "HTTP listening endpoint",
	)
	flagHTTPSListenAddress = flag.String(
		"https-listen-address", ":8443", "HTTPS listening endpoint",
	)
	flagTLSCert = flag.String(
		"tls-cert", "cert.pem", "path to the TLS certificate file to use",
	)
	flagTLSKey = flag.String(
		"tls-key", "key.pem", "path to the TLS key to use",
	)
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
			*flagHTTPSListenAddress, *flagTLSCert, *flagTLSKey, rootHandler,
		), "Can't start HTTPS server")
	}()
	rtx.Must(http.ListenAndServe(
		*flagHTTPListenAddress, rootHandler), "Can't start HTTP server")
}
