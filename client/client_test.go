package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	locatev2 "github.com/m-lab/locate/api/v2"
	"github.com/neubot/dash/model"
)

const (
	softwareName    = "dash-client-go-test"
	softwareVersion = "0.0.1"
)

func TestClientNegotiate(t *testing.T) {
	t.Run("json.Marshal failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.NewRequest failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPNewRequest = func(
			method string, url string, body io.Reader,
		) (*http.Request, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("io.ReadAll failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		client.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Invalid JSON or not authorized", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(strings.NewReader(`{
					"Authorization": "0xdeadbeef",
					"Unchoked": 1
				}`)),
			}, nil
		}
		_, err := client.negotiate(context.Background(), &url.URL{})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientDownload(t *testing.T) {
	t.Run("http.NewRequest failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPNewRequest = func(
			method string, url string, body io.Reader,
		) (*http.Request, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current, &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current, &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current, &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("io.ReadAll failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		client.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current, &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current, &url.URL{})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientCollect(t *testing.T) {
	t.Run("json.Marshal failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.NewRequest failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPNewRequest = func(
			method string, url string, body io.Reader,
		) (*http.Request, error) {
			return nil, errors.New("Mocked error")
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("io.ReadAll failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		client.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("[]")),
			}, nil
		}
		err := client.collect(context.Background(), "abc", &url.URL{})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientLoop(t *testing.T) {
	t.Run("negotiate failure", func(t *testing.T) {
		ch := make(chan model.ClientResults)
		client := New(softwareName, softwareVersion)
		client.deps.Negotiate = func(ctx context.Context, negotiateURL *url.URL) (model.NegotiateResponse, error) {
			return model.NegotiateResponse{}, errors.New("Mocked error")
		}
		client.loop(context.Background(), ch, &url.URL{})
		if client.err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("download failure", func(t *testing.T) {
		ch := make(chan model.ClientResults)
		client := New(softwareName, softwareVersion)
		client.deps.Negotiate = func(ctx context.Context, negotiateURL *url.URL) (model.NegotiateResponse, error) {
			return model.NegotiateResponse{}, nil
		}
		client.deps.Download = func(
			ctx context.Context, authorization string,
			current *model.ClientResults, negotiateURL *url.URL,
		) error {
			return errors.New("Mocked error")
		}
		client.loop(context.Background(), ch, &url.URL{})
		if client.err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("collect failure", func(t *testing.T) {
		ch := make(chan model.ClientResults)
		client := New(softwareName, softwareVersion)
		client.deps.Negotiate = func(ctx context.Context, negotiateURL *url.URL) (model.NegotiateResponse, error) {
			return model.NegotiateResponse{}, nil
		}
		client.deps.Download = func(
			ctx context.Context, authorization string,
			current *model.ClientResults, negotiateURL *url.URL,
		) error {
			return nil
		}
		client.deps.Collect = func(ctx context.Context, authorization string, negotiateURL *url.URL) error {
			return errors.New("Mocked error")
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				// drain channel
			}
		}()
		client.loop(context.Background(), ch, &url.URL{})
		if client.err == nil {
			t.Fatal("Expected an error here")
		}
		wg.Wait() // make sure we really terminate
	})
}

type failingLocator struct{}

// Nearest implements locator.
func (f *failingLocator) Nearest(ctx context.Context, service string) ([]locatev2.Target, error) {
	return nil, errors.New("mocked error")
}

func TestClientStartDownload(t *testing.T) {
	t.Run("mlabns failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.Locator = &failingLocator{}
		ch, err := client.StartDownload(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
		if ch != nil {
			t.Fatal("Expected nil channel here")
		}
	})

	t.Run("common case", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.Loop = func(ctx context.Context, ch chan<- model.ClientResults, negotiateURL *url.URL) {
			close(ch)
		}
		ch, err := client.StartDownload(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for range ch {
			// drain channel
		}
	})
}
