// Package mockable contains mockable code
package mockable

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// Dependencies contains the mockable dependencies. Usually you don't care
// about this code, except when you are running tests.
type Dependencies struct {
	// GzipNewWriterLevel is like NewWriter but specifies the compression
	// level instead of assuming DefaultCompression.
	GzipNewWriterLevel func(w io.Writer, level int) (*gzip.Writer, error)

	// HTTPNewRequest returns a new Request given a
	// method, URL, and optional body.
	HTTPNewRequest func(method, url string, body io.Reader) (*http.Request, error)

	// JSONMarshal returns the JSON encoding of v.
	JSONMarshal func(v interface{}) ([]byte, error)

	// OSMkdirAll creates a directory named path, along with any
	// necessary parents, and returns nil, or else returns an error.
	OSMkdirAll func(path string, perm os.FileMode) error

	// OpenFile is the generalized open call; most users will
	// use Open or Create instead.
	OSOpenFile func(name string, flag int, perm os.FileMode) (*os.File, error)

	// RandRead generates len(p) random bytes from the default
	// rand.Source and writes them into p.
	RandRead func(p []byte) (n int, err error)

	// UUIDNewRandom returns a Random (Version 4) UUID.
	UUIDNewRandom func() (uuid.UUID, error)
}

// NewDependencies creates a new instance of the dependencies struct
// where all dependencies have their default value.
func NewDependencies() Dependencies {
	return Dependencies{
		GzipNewWriterLevel: gzip.NewWriterLevel,
		HTTPNewRequest:     http.NewRequest,
		JSONMarshal:        json.Marshal,
		OSMkdirAll:         os.MkdirAll,
		OSOpenFile:         os.OpenFile,
		RandRead:           rand.Read,
		UUIDNewRandom:      uuid.NewRandom,
	}
}
