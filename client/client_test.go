package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

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
		_, err := client.negotiate(context.Background())
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
		_, err := client.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := client.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
			}, nil
		}
		_, err := client.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		client.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := client.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		_, err := client.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Invalid JSON or not authorized", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil
		}
		_, err := client.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"Authorization": "0xdeadbeef",
					"Unchoked": 1
				}`)),
			}, nil
		}
		_, err := client.negotiate(context.Background())
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
		err := client.download(context.Background(), "abc", current)
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
		err := client.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
			}, nil
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		client.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		current := new(model.ClientResults)
		err := client.download(context.Background(), "abc", current)
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
		err := client.collect(context.Background(), "abc")
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
		err := client.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		err := client.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
			}, nil
		}
		err := client.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		client.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		err := client.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		err := client.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("[]")),
			}, nil
		}
		err := client.collect(context.Background(), "abc")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientLoop(t *testing.T) {
	t.Run("negotiate failure", func(t *testing.T) {
		ch := make(chan model.ClientResults)
		client := New(softwareName, softwareVersion)
		client.deps.Negotiate = func(ctx context.Context) (model.NegotiateResponse, error) {
			return model.NegotiateResponse{}, errors.New("Mocked error")
		}
		client.loop(context.Background(), ch)
		if client.err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("download failure", func(t *testing.T) {
		ch := make(chan model.ClientResults)
		client := New(softwareName, softwareVersion)
		client.deps.Negotiate = func(ctx context.Context) (model.NegotiateResponse, error) {
			return model.NegotiateResponse{}, nil
		}
		client.deps.Download = func(
			ctx context.Context, authorization string, current *model.ClientResults,
		) error {
			return errors.New("Mocked error")
		}
		client.loop(context.Background(), ch)
		if client.err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("collect failure", func(t *testing.T) {
		ch := make(chan model.ClientResults)
		client := New(softwareName, softwareVersion)
		client.deps.Negotiate = func(ctx context.Context) (model.NegotiateResponse, error) {
			return model.NegotiateResponse{}, nil
		}
		client.deps.Download = func(
			ctx context.Context, authorization string, current *model.ClientResults,
		) error {
			return nil
		}
		client.deps.Collect = func(ctx context.Context, authorization string) error {
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
		client.loop(context.Background(), ch)
		if client.err == nil {
			t.Fatal("Expected an error here")
		}
		wg.Wait() // make sure we really terminate
	})
}

func TestClientStartDownload(t *testing.T) {
	t.Run("mlabns failure", func(t *testing.T) {
		client := New(softwareName, softwareVersion)
		client.deps.Locate = func(ctx context.Context) (string, error) {
			return "", errors.New("Mocked error")
		}
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
		client.deps.Loop = func(ctx context.Context, ch chan<- model.ClientResults) {
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
