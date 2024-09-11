// Package client implements the DASH client
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/m-lab/locate/api/locate"
	locatev2 "github.com/m-lab/locate/api/v2"
	"github.com/neubot/dash/internal"
	"github.com/neubot/dash/model"
	"github.com/neubot/dash/spec"
)

const (
	// libraryName is the name of this library
	libraryName = "neubot-dash"

	// libraryVersion is the version of this library.
	libraryVersion = "0.4.3"

	// magicVersion is a magic number that identifies in a unique
	// way this implementation of DASH. 0.007xxxyyy is Measurement
	// Kit. Values lower than that are Neubot.
	magicVersion = "0.008000000"
)

var (
	// ErrServerBusy is returned when the Neubot server is busy.
	ErrServerBusy = errors.New("server busy; try again later")

	// errHTTPRequestFailed is returned when an HTTP request fails.
	errHTTPRequestFailed = errors.New("HTTP request failed")
)

// locator is an interface used to locate a server.
type locator interface {
	Nearest(ctx context.Context, service string) ([]locatev2.Target, error)
}

// dependencies contains mockable dependencies to test the client
type dependencies struct {
	// Collect allows to override the method performing the collect phase.
	Collect func(ctx context.Context, authorization string, negotiateURL *url.URL) error

	// Download allows to override the method performing the download phase.
	Download func(
		ctx context.Context, authorization string,
		current *model.ClientResults,
		negotiateURL *url.URL) error

	// HTTPClientDo allows to override calling the [*http.Client.Do].
	HTTPClientDo func(req *http.Request) (*http.Response, error)

	// HTTPNewRequest allows to override calling [http.NewRequest].
	HTTPNewRequest func(method, url string, body io.Reader) (*http.Request, error)

	// IOReadAll allows to override calling [io.ReadAll].
	IOReadAll func(r io.Reader) ([]byte, error)

	// JSONUnmarshal allows to override calling [json.Unmarshal].
	JSONMarshal func(v interface{}) ([]byte, error)

	// Locator allows to override the [locator] to use.
	Locator locator

	// Loop allows to override the function running the DASH client loop.
	Loop func(ctx context.Context, ch chan<- model.ClientResults, negotiateURL *url.URL)

	// Negotiate allows to override the method performing the negotiate phase.
	Negotiate func(ctx context.Context, negotiateURL *url.URL) (model.NegotiateResponse, error)
}

// Client is a DASH client. The zero value of this structure is
// invalid. Use NewClient to correctly initialize the fields.
type Client struct {
	// ClientName is the name of the client application. This field is
	// initialized by the NewClient constructor.
	ClientName string

	// ClientVersion is the version of the client application. This field is
	// initialized by the NewClient constructor.
	ClientVersion string

	// FQDN is the server of the server to use. If the FQDN is not
	// specified, we use m-lab/locate/v2 to discover a server.
	FQDN string

	// HTTPClient is the HTTP client used by this implementation. This field
	// is initialized by the NewClient to http.DefaultClient.
	HTTPClient *http.Client

	// Logger is the logger to use. This field is initialized by the
	// NewClient constructor to a do-nothing logger.
	Logger model.Logger

	// Scheme is the protocol scheme to use. By default NewClient configures
	// it to "https", but you can override it to "http".
	Scheme string

	// begin is when the test started.
	begin time.Time

	// clientResults contains results collected by the client.
	clientResults []model.ClientResults

	// deps contains the mockable dependencies.
	deps dependencies

	// err is the overall error that occurred.
	err error

	// numIterations is the number of iterations to run.
	numIterations int64

	// serverResults contains the server results.
	serverResults []model.ServerResults

	// userAgent is the user-agent HTTP header to use.
	userAgent string
}

func makeUserAgent(clientName, clientVersion string) string {
	return clientName + "/" + clientVersion + " " + libraryName + "/" + libraryVersion
}

func (c *Client) httpClientDo(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}

// New creates a new Client instance using the specified
// client application name and version.
func New(clientName, clientVersion string) (client *Client) {
	ua := makeUserAgent(clientName, clientVersion)
	client = &Client{
		ClientName:    clientName,
		ClientVersion: clientVersion,
		FQDN:          "", // user specified and defaults to empty
		HTTPClient:    http.DefaultClient,
		Logger:        internal.NoLogger{},
		Scheme:        "https",
		begin:         time.Now(),
		clientResults: []model.ClientResults{},
		deps:          dependencies{}, // initialized below
		err:           nil,
		numIterations: 15,
		serverResults: []model.ServerResults{},
		userAgent:     ua,
	}
	client.deps = dependencies{
		Collect:        client.collect,
		Download:       client.download,
		HTTPClientDo:   client.httpClientDo,
		HTTPNewRequest: http.NewRequest,
		IOReadAll:      io.ReadAll,
		JSONMarshal:    json.Marshal,
		Locator:        locate.NewClient(ua),
		Loop:           client.loop,
		Negotiate:      client.negotiate,
	}
	return
}

// negotiate is the preliminary phase of Neubot experiment where we connect
// to the server, negotiate test parameters, and obtain an authorization
// token that will be used by us and by the server to identify this experiment.
func (c *Client) negotiate(
	ctx context.Context,
	negotiateURL *url.URL,
) (model.NegotiateResponse, error) {
	// 1. create the HTTP request
	//
	// TODO(bassosimone): use http.NewRequestWithContext
	var negotiateResponse model.NegotiateResponse
	data, err := c.deps.JSONMarshal(model.NegotiateRequest{
		DASHRates: spec.DefaultRates,
	})
	if err != nil {
		return negotiateResponse, err
	}
	c.Logger.Debugf("dash: body: %s", string(data))
	req, err := c.deps.HTTPNewRequest("POST", negotiateURL.String(), bytes.NewReader(data))
	if err != nil {
		return negotiateResponse, err
	}
	c.Logger.Debugf("dash: POST %s", negotiateURL.String())
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "")
	req = req.WithContext(ctx)

	// 2. send the request and receive the response headers
	resp, err := c.deps.HTTPClientDo(req)
	if err != nil {
		return negotiateResponse, err
	}
	defer resp.Body.Close()

	// 3. handle the case where the status code indicates failure
	c.Logger.Debugf("dash: StatusCode: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return negotiateResponse, errHTTPRequestFailed
	}

	// 4. read the raw response body
	//
	// TODO(bassosimone):
	//
	// a) protect against arbitrarily large bodies
	//
	// b) make sure the context can still interrupt a client otherwise
	// with some amount of interference, we'll block here forever
	data, err = c.deps.IOReadAll(resp.Body)
	if err != nil {
		return negotiateResponse, err
	}

	// 5. parse the response body
	c.Logger.Debugf("dash: body: %s", string(data))
	err = json.Unmarshal(data, &negotiateResponse)
	if err != nil {
		return negotiateResponse, err
	}

	// 6. make sure that the server isn't busy
	//
	// Implementation oddity: Neubot is using an integer rather than a
	// boolean for the unchoked, with obvious semantics. I wonder why
	// I choose an integer over a boolean, given that Python does have
	// support for booleans. I don't remember 🤷.
	if negotiateResponse.Authorization == "" || negotiateResponse.Unchoked == 0 {
		return negotiateResponse, ErrServerBusy
	}
	c.Logger.Debugf("dash: authorization: %s", negotiateResponse.Authorization)
	return negotiateResponse, nil
}

// makeDownloadURL makes the download URL from the negotiate URL.
func makeDownloadURL(negotiateURL *url.URL, path string) *url.URL {
	return &url.URL{
		Scheme: negotiateURL.Scheme,
		Host:   negotiateURL.Host,
		Path:   path,
	}
}

// download implements the DASH test proper. We compute the number of bytes
// to request given the current rate, download the fake DASH segment, and
// then we return the measured performance of this segment to the caller. This
// is repeated several times to emulate downloading part of a video.
func (c *Client) download(
	ctx context.Context,
	authorization string,
	current *model.ClientResults,
	negotiateURL *url.URL,
) error {
	// 1. create the HTTP request
	//
	// TODO(bassosimone): use http.NewRequestWithContext
	nbytes := (current.Rate * 1000 * current.ElapsedTarget) >> 3
	URL := makeDownloadURL(negotiateURL, fmt.Sprintf("%s%d", spec.DownloadPath, nbytes))
	req, err := c.deps.HTTPNewRequest("GET", URL.String(), nil)
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: GET %s", URL.String())
	current.ServerURL = URL.String()
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", authorization)
	req = req.WithContext(ctx)
	savedTicks := time.Now()

	// 2. send the request and receive the response headers
	resp, err := c.deps.HTTPClientDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 3. handle the case where the status code indicates failure
	c.Logger.Debugf("dash: StatusCode: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return errHTTPRequestFailed
	}

	// 4. read the raw response body
	//
	// TODO(bassosimone):
	//
	// a) protect against arbitrarily large bodies
	//
	// b) make sure the context can still interrupt a client otherwise
	// with some amount of interference, we'll block here forever
	data, err := c.deps.IOReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 5. compute performance metrics and update current
	//
	// Implementation note: MK contains a comment that says that Neubot uses
	// the elapsed time since when we start receiving the response but it
	// turns out that Neubot and MK do the same. So, we do what they do. At
	// the same time, we are currently not able to include the overhead that
	// is caused by HTTP headers etc. So, we're a bit less precise.
	current.Elapsed = time.Since(savedTicks).Seconds()
	current.Received = int64(len(data))
	current.RequestTicks = savedTicks.Sub(c.begin).Seconds()
	current.Timestamp = time.Now().Unix()

	//c.Logger.Debugf("dash: current: %+v", current) /* for debugging */
	return nil
}

// makeCollectURL makes the collect URL from the negotiate URL.
func makeCollectURL(negotiateURL *url.URL) *url.URL {
	return &url.URL{
		Scheme: negotiateURL.Scheme,
		Host:   negotiateURL.Host,
		Path:   spec.CollectPath,
	}
}

// collect is the final phase of the test. We send to the server what we
// measured and we receive back what it has measured.
func (c *Client) collect(
	ctx context.Context,
	authorization string,
	negotiateURL *url.URL,
) error {
	// 1. create the HTTP request including the JSON request body
	//
	// TODO(bassosimone): our request constructor should use http.NewRequestWithContext
	// such that we don't actually need to set the context as a separate operation
	data, err := c.deps.JSONMarshal(c.clientResults)
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: body: %s", string(data))
	URL := makeCollectURL(negotiateURL)
	req, err := c.deps.HTTPNewRequest("POST", URL.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.Logger.Debugf("dash: POST %s", URL.String())
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	req = req.WithContext(ctx)

	// 2. send the request and receive the corresponding response headers
	resp, err := c.deps.HTTPClientDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 3. handle the case where the status code indicates failure
	c.Logger.Debugf("dash: StatusCode: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return errHTTPRequestFailed
	}

	// 4. read the raw response body
	//
	// TODO(bassosimone):
	//
	// a) protect against arbitrarily large bodies
	//
	// b) make sure the context can still interrupt a client otherwise
	// with some amount of interference, we'll block here forever
	data, err = c.deps.IOReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 5. parse the response body
	//
	// Implementation note: historically this client did never care
	// about saving the response body and we're still doing this
	c.Logger.Debugf("dash: body: %s", string(data))
	return json.Unmarshal(data, &c.serverResults)
}

// loop is the main loop of the DASH test. It performs negotiation, the test
// proper, and then collection. It posts interim results on |ch|.
func (c *Client) loop(
	ctx context.Context,
	ch chan<- model.ClientResults,
	negotiateURL *url.URL,
) {
	// 1. make sure we close the channel when done
	defer close(ch)

	// 2. negotiate an authorization token with the server
	//
	// Implementation note: we will soon refactor the server to eliminate the
	// possiblity of keeping clients in queue. For this reason it's becoming
	// increasingly less important to loop waiting for the ready signal. Hence
	// if the server is busy, we just return a well known error.
	var negotiateResponse model.NegotiateResponse
	negotiateResponse, c.err = c.deps.Negotiate(ctx, negotiateURL)
	if c.err != nil {
		return
	}

	// 3. run the measurement loop proper
	//
	// Note: according to a comment in MK sources 3000 kbit/s was the
	// minimum speed recommended by Netflix for SD quality in 2017.
	//
	// See: <https://help.netflix.com/en/node/306>.
	const initialBitrate = 3000
	current := model.ClientResults{
		ElapsedTarget: 2,
		Platform:      runtime.GOOS,
		Rate:          initialBitrate,
		RealAddress:   negotiateResponse.RealAddress,
		Version:       magicVersion,
	}
	for current.Iteration < c.numIterations {
		c.err = c.deps.Download(ctx, negotiateResponse.Authorization, &current, negotiateURL)
		if c.err != nil {
			return
		}
		c.clientResults = append(c.clientResults, current)
		ch <- current
		current.Iteration++
		speed := float64(current.Received) / float64(current.Elapsed)
		speed *= 8.0    // to bits per second
		speed /= 1000.0 // to kbit/s
		current.Rate = int64(speed)
	}

	// 4. submit the measurement results
	c.err = c.deps.Collect(ctx, negotiateResponse.Authorization, negotiateURL)
}

// StartDownload starts the DASH download. It returns a channel where
// client measurements are posted, or an error. This function will only
// fail if we cannot even initiate the experiment. If you see some
// results on the returned channel, then maybe it means the experiment
// has somehow worked. You can see if there has been any error during
// the experiment by using the Error function.
func (c *Client) StartDownload(ctx context.Context) (<-chan model.ClientResults, error) {

	// 1. use the provided -fqdn or use m-lab/locate/v2
	var negotiateURL *url.URL
	switch {

	// 1.1: the user manually specified the server -fqdn
	case c.FQDN != "":
		negotiateURL.Scheme = c.Scheme
		negotiateURL.Host = c.FQDN
		negotiateURL.Path = spec.NegotiatePath

	// 1.2: we're going to use m-lab/locate/v2 for discovering the server
	default:
		c.Logger.Debug("dash: discovering server with locate v2")

		targets, err := c.deps.Locator.Nearest(ctx, "neubot/dash")
		if err != nil {
			return nil, err
		}
		if len(targets) < 1 {
			return nil, errors.New("no targets")
		}

		URL := targets[0].URLs["https:///negotiate/dash"]
		parsed, err := url.Parse(URL)
		if err != nil {
			return nil, err
		}

		negotiateURL = parsed
	}

	// 2. check for context being canceled
	//
	// this check is useful to write better tests
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 3. run the client loop and return the resulting channel
	c.Logger.Debugf("dash: using server: %v", negotiateURL)
	ch := make(chan model.ClientResults)
	go c.deps.Loop(ctx, ch, negotiateURL)
	return ch, nil
}

// Error returns the error that occurred during the test, if any. A nil
// return value means that all was good. A returned error does not however
// necessarily mean that all was bad; you may have _some_ data.
func (c *Client) Error() error {
	// TODO(bassosimone): I am not convinced about writing into the
	// err field without any locking and should double check this
	return c.err
}

// ServerResults returns the results of the experiment collected by the
// server. In case Error() returns non nil, this function will typically
// return an empty slice to the caller.
func (c *Client) ServerResults() []model.ServerResults {
	return c.serverResults
}
