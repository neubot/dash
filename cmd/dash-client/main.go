// dash-client is the dash command line client.
//
// Usage:
//
//    dash-client [-hostname <name>] [-timeout <string>]
//
// The `-hostname <name>` flag specifies to use the `name` hostname for
// performing the dash test. The default is to autodiscover a suitable
// server by using Measurement Lab's locate service.
//
// The `-timeout <string>` flag specifies the time after which the
// whole test is interrupted. The `<string>` is a string suitable to
// be passed to time.ParseDuration, e.g., "15s". The default is a large
// enough value that should be suitable for common conditions.
//
// Additionally, passing any unrecognized flag, such as `-help`, will
// cause dash-client to print a brief help message.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/neubot/dash/client"
)

const (
	clientName     = "dash-client-go"
	clientVersion  = "0.1.0"
	defaultTimeout = 55 * time.Second
)

var (
	flagHostname = flag.String("hostname", "", "optional ndt7 server hostname")
	flagTimeout  = flag.Duration(
		"timeout", defaultTimeout, "time after which the test is aborted")
)

func main() {
	log.SetLevel(log.DebugLevel)
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), *flagTimeout)
	defer cancel()
	client := client.New(clientName, clientVersion)
	client.Logger = log.Log
	client.FQDN = *flagHostname
	ch, err := client.StartDownload(ctx)
	if err != nil {
		log.WithError(err).Fatal("StartDownload failed")
	}
	for results := range ch {
		data, err := json.Marshal(results)
		if err != nil {
			log.WithError(err).Fatal("json.Marshal failed")
		}
		fmt.Printf("%s\n", string(data))
	}
	if client.Error() != nil {
		log.WithError(client.Error()).Fatal("the download failed")
	}
	data, err := json.Marshal(client.ServerResults())
	if err != nil {
		log.WithError(err).Fatal("json.Marshal failed")
	}
	fmt.Printf("%s\n", string(data))
}
