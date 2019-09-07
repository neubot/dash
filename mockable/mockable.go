// Package mockable contains mockable code
package mockable

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

// Dependencies contains the mockable dependencies. Usually you don't care
// about this code, except when you are running tests.
type Dependencies struct {
	// HTTPNewRequest returns a new Request given a
	// method, URL, and optional body.
	HTTPNewRequest func(method, url string, body io.Reader) (*http.Request, error)

	// UUIDNewRandom returns a Random (Version 4) UUID.
	UUIDNewRandom func() (uuid.UUID, error)

	// JSONMarshal returns the JSON encoding of v.
	JSONMarshal func(v interface{}) ([]byte, error)

	// RandRead generates len(p) random bytes from the default
	// rand.Source and writes them into p.
	RandRead func(p []byte) (n int, err error)
}

// NewDependencies creates a new instance of the dependencies struct
// where all dependencies have their default value.
func NewDependencies() Dependencies {
	return Dependencies{
		HTTPNewRequest: http.NewRequest,
		UUIDNewRandom:  uuid.NewRandom,
		JSONMarshal:    json.Marshal,
		RandRead:       rand.Read,
	}
}
