// Package common contains common code
package common

const (
	// CurrentServerSchemaVersion is the version of the server schema that
	// will be adopted by this implementation. Version 3 is the one that is
	// Neubot uses. We needed to bump the version because Web100 is not on
	// M-Lab anymore and hence we need to make a breaking change.
	CurrentServerSchemaVersion = 4

	// NegotiatePath is the URL path used to negotiate
	NegotiatePath = "/negotiate/dash"

	// DownloadPathNoTrailingSlash is like DownloadPath but has no
	// trailing slash. For historical reasons we also need to handle
	// this path in addition to DownloadPath.
	DownloadPathNoTrailingSlash = "/dash/download"

	// DownloadPath is the URL path used to request DASH segments. You can
	// append to this path an integer indicating how many bytes you would like
	// the server to send you as part of the next chunk.
	DownloadPath = DownloadPathNoTrailingSlash + "/"

	// CollectPath is the URL path used to collect
	CollectPath = "/collect/dash"
)

// ClientResults contains the results measured by the client. This data
// structure is sent to the server in the collection phase.
//
// All the fields listed here are part of the original specification
// of DASH, except ServerURL, added in MK v0.10.6.
type ClientResults struct {
	ConnectTime     float64 `json:"connect_time"`
	DeltaSysTime    int64   `json:"delta_sys_time"`
	DeltaUserTime   int64   `json:"delta_user_time"`
	Elapsed         float64 `json:"elapsed"`
	ElapsedTarget   int64   `json:"elapsed_target"`
	InternalAddress string  `json:"internal_address"`
	Iteration       int64   `json:"iteration"`
	Platform        string  `json:"platform"`
	Rate            int64   `json:"rate"`
	RealAddress     string  `json:"real_address"`
	Received        int64   `json:"received"`
	RemoteAddress   string  `json:"remote_address"`
	RequestTicks    float64 `json:"request_ticks"`
	ServerURL       string  `json:"server_url"`
	Timestamp       int64   `json:"timestamp"`
	UUID            string  `json:"uuid"`
	Version         string  `json:"version"`
}

// ServerResults contains the server results. This data structure is sent
// to the client during the collection phase of DASH.
type ServerResults struct {
	Iteration int64   `json:"iteration"`
	Ticks     float64 `json:"ticks"`
	Timestamp int64   `json:"timestamp"`
}

// ServerSchema is the data format traditionally used by the
// original Neubot server for DASH experiments.
type ServerSchema struct {
	Client              []ClientResults `json:"client"`
	ServerSchemaVersion int             `json:"srvr_schema_version"`
	ServerTimestamp     int64           `json:"srvr_timestamp"`
	Server              []ServerResults `json:"server"`
}

// NegotiateRequest contains the request of negotiation
type NegotiateRequest struct {
	DASHRates []int64 `json:"dash_rates"`
}

// NegotiateResponse contains the response of negotiation
type NegotiateResponse struct {
	Authorization string `json:"authorization"`
	QueuePos      int64  `json:"queue_pos"`
	RealAddress   string `json:"real_address"`
	Unchoked      int    `json:"unchoked"`
}

// DefaultRates contains the default DASH rates in kbit/s.
var DefaultRates = []int64{
	100,
	150,
	200,
	250,
	300,
	400,
	500,
	700,
	900,
	1200,
	1500,
	2000,
	2500,
	3000,
	4000,
	5000,
	6000,
	7000,
	10000,
	20000,
}

// Logger defines the common interface that a logger should have. It is
// out of the box compatible with `log.Log` in `apex/log`.
//
// This interface is copied from github.com/m-lab/ndt7-client-go.
type Logger interface {
	// Debug emits a debug message.
	Debug(msg string)

	// Debugf formats and emits a debug message.
	Debugf(format string, v ...interface{})

	// Info emits an informational message.
	Info(msg string)

	// Infof format and emits an informational message.
	Infof(format string, v ...interface{})

	// Warn emits a warning message.
	Warn(msg string)

	// Warnf formats and emits a warning message.
	Warnf(format string, v ...interface{})
}
