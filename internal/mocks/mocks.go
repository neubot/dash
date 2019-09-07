// Package mocks contains mocks
package mocks

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrMocked is returned by mocked code as a sentinel failure
	ErrMocked = errors.New("Mocked error")
)

// HTTPRoundTripInfo contains info about a round trip
type HTTPRoundTripInfo struct {
	Error    error
	Response *http.Response
	URL      string
}

// HTTPTransport is a mocked HTTP Transport
type HTTPTransport struct {
	Info       []HTTPRoundTripInfo
	currentIdx int
}

// RoundTrip performs an HTTP request and returns the response.
func (r *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// Note: it's okay to panic if we've not configured the proper number of URLs
	// since this is a very noisy way to tell us that tests are broken.
	info := r.Info[r.currentIdx]
	r.currentIdx++
	if req.URL.String() != info.URL {
		panic(fmt.Sprintf("Unexpected URL: %s", req.URL.String()))
	}
	resp, err = info.Response, info.Error
	return
}

// NewHTTPClient returns a new HTTP client that will behave as
// specified by the ...infos arguments.
func NewHTTPClient(infos ...HTTPRoundTripInfo) *http.Client {
	return &http.Client{
		Transport: &HTTPTransport{
			Info: infos,
		},
	}
}

// NewHTTPRoundTripFailure is a convenience function to create a
// HTTPRoundTripInfo with a specific failure and URL.
func NewHTTPRoundTripFailure(URL string) HTTPRoundTripInfo {
	return HTTPRoundTripInfo{
		URL:      URL,
		Error:    ErrMocked,
		Response: nil,
	}
}

// ProgrammableReader is a reader that you can program
type ProgrammableReader struct {
	Reader io.Reader
	Err    error
}

// Read attempts to read from the FailingReader instance
func (pr ProgrammableReader) Read(d []byte) (int, error) {
	if pr.Err != nil {
		return 0, pr.Err
	}
	return pr.Reader.Read(d)
}
