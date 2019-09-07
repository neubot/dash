package mockable

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

var (
	// HTTPNewRequest returns a new Request given
	// a method, URL, and optional body.
	HTTPNewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		return http.NewRequest(method, url, body)
	}

	// NewRandomUUID returns a Random (Version 4) UUID.
	NewRandomUUID = func() (uuid.UUID, error) {
		return uuid.NewRandom()
	}

	// MarshalJSON returns the JSON encoding of v.
	MarshalJSON = func(v interface{}) ([]byte, error) {
		return json.Marshal(v)
	}

	// RandRead generates len(p) random bytes from the default
	// rand.Source and writes them into p.
	RandRead = func(p []byte) (n int, err error) {
		return rand.Read(p)
	}
)
