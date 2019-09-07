package mocks_test

import (
	"net/http"
	"testing"

	"github.com/neubot/dash/internal/mocks"
)

func TestHTTPTransportPanicsIfURLDiffers(t *testing.T) {
	client := mocks.NewHTTPClient(
		mocks.HTTPRoundTripInfo{
			URL: "http://www.example.com",
		},
	)
	request, err := http.NewRequest("GET", "http://www.example.org", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("The code did not panic")
		}
	}()
	client.Do(request)
}
