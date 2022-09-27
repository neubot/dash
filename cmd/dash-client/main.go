// dash-client is the dash command line client.
//
// Usage:
//
//    dash-client -y [-hostname <name>] [-timeout <string>] [-scheme <scheme>]
//
// The `-y` flag indicates you have read the data policy and accept it.
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
	"os"
	"time"

	"github.com/apex/log"
	"github.com/m-lab/go/flagx"
	"github.com/m-lab/go/rtx"
	"github.com/neubot/dash/client"
)

const (
	clientName     = "dash-client-go"
	clientVersion  = "0.4.3"
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
	flagY = flag.Bool("y", false, "I have read and accept the privacy policy at https://github.com/neubot/dash/blob/master/PRIVACY.md")
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
	if !*flagY {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Please, read the privacy policy at https://github.com/neubot/dash/blob/master/PRIVACY.md.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "If you accept the privacy policy, rerun adding the `-y` flag to the command line.\n")
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}
	client := client.New(clientName, clientVersion)
	client.Logger = log.Log
	client.FQDN = *flagHostname
	client.Scheme = flagScheme.Value
	return realmain(ctx, client, *flagTimeout, nil)
}

func fmain(f func(context.Context) error, e func(error, string, ...interface{})) {
	if err := f(context.Background()); err != nil {
		e(err, "DASH experiment failed")
	}
}

var defaultMain = internalmain // testability

func main() {
	fmain(defaultMain, rtx.Must)
}
