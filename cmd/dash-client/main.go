// dash-client is the dash command line client.
//
// Usage:
//
//    dash-client [-hostname <name>] [-timeout <string>] [-scheme <scheme>]
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
// The `-scheme <scheme>` flag allows to override the default scheme
// used for the test, i.e. "http". All DASH servers support that,
// future versions of the Go server will support "https".
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
	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
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
	flagScheme = flagx.Enum{
		Options: []string{"https", "http"},
		Value:   "https",
	}
)

func init() {
	flag.Var(
		&flagScheme,
		"scheme",
		`Protocol scheme to use: either "https" (the default) or "http"`,
	)
}

func realmain(ctx context.Context, client *client.Client, timeout time.Duration, onresult func()) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ch, err := client.StartDownload(ctx)
	if err != nil {
		return err
	}
	for results := range ch {
		if onresult != nil {
			onresult() // this is an hook that we use for testing
		}
		data, err := json.Marshal(results)
		rtx.PanicOnError(err, "json.Marshal should not fail")
		fmt.Printf("%s\n", string(data))
	}
	if client.Error() != nil {
		return client.Error()
	}
	data, err := json.Marshal(client.ServerResults())
	rtx.PanicOnError(err, "json.Marshal should not fail")
	fmt.Printf("%s\n", string(data))
	return nil
}

func init() {
	log.SetLevel(log.DebugLevel) // needs to run exactly once
}

func internalmain(ctx context.Context) error {
	flag.Parse()
	client := client.New(clientName, clientVersion)
	client.Logger = log.Log
	client.FQDN = *flagHostname
	client.Scheme = flagScheme.Value
	return realmain(ctx, client, *flagTimeout, nil)
}

func main() {
	if err := internalmain(context.Background()); err != nil {
		log.WithError(err).Fatal("DASH experiment failed")
	}
}
