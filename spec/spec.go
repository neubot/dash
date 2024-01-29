// Package spec contains constants
package spec

const (
	// CurrentServerSchemaVersion is the version of the server schema that
	// will be adopted by this implementation. Version 3 is the one that is
	// Neubot uses. We needed to bump the version because Web100 is not on
	// M-Lab anymore and hence we need to make a breaking change.
	CurrentServerSchemaVersion = 4

	// NegotiatePath is the URL path used to negotiate. We use /negotiate/dash
	// rather than /dash/negotiate for historical reasons. Neubot used to
	// handle all requests for negotiation by handling the /negotiate prefix
	// and routing to the proper experiment.
	NegotiatePath = "/negotiate/dash"

	// DownloadPathNoTrailingSlash is like DownloadPath but has no
	// trailing slash. For historical reasons we also need to handle
	// this path in addition to DownloadPath.
	DownloadPathNoTrailingSlash = "/dash/download"

	// DownloadPath is the URL path used to request DASH segments. You can
	// append to this path an integer indicating how many bytes you would like
	// the server to send you as part of the next chunk.
	DownloadPath = DownloadPathNoTrailingSlash + "/"

	// CollectPath is the URL path used to collect. We use /collect/dash
	// rather than /dash/collect for historical reasons. Neubot used to
	// handle all requests for collection by handling the /collect prefix
	// and routing to the proper experiment.
	CollectPath = "/collect/dash"
)

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
