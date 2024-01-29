package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/google/uuid"
	"github.com/neubot/dash/model"
)

func TestServerNegotiate(t *testing.T) {
	t.Run("net.SplitHostPort failure", func(t *testing.T) {
		handler := NewHandler("")
		req := new(http.Request)
		w := httptest.NewRecorder()
		handler.negotiate(w, req)
		resp := w.Result()
		if resp.StatusCode != 500 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("uuid.NewRandom failure", func(t *testing.T) {
		handler := NewHandler("")
		handler.deps.UUIDNewRandom = func() (uuid.UUID, error) {
			return uuid.UUID{}, errors.New("Mocked error")
		}
		req := new(http.Request)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()
		handler.negotiate(w, req)
		resp := w.Result()
		if resp.StatusCode != 500 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("json.Marshal failure", func(t *testing.T) {
		handler := NewHandler("")
		handler.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		req := new(http.Request)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()
		handler.negotiate(w, req)
		resp := w.Result()
		if resp.StatusCode != 500 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("common case", func(t *testing.T) {
		handler := NewHandler("")
		req := new(http.Request)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()
		handler.negotiate(w, req)
		resp := w.Result()
		if resp.StatusCode != 200 {
			t.Fatal("Expected different status code")
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		var msg model.NegotiateResponse
		err = json.Unmarshal(data, &msg)
		if err != nil {
			t.Fatal(err)
		}
		if len(msg.Authorization) <= 0 {
			t.Fatal("Authorization is empty")
		}
		if msg.QueuePos != 0 {
			t.Fatal("QueuePos is nonzero")
		}
		if msg.RealAddress != "127.0.0.1" {
			t.Fatal("RealAddress is wrong")
		}
		if msg.Unchoked != 1 {
			t.Fatal("Unchoked is different from one")
		}
		if handler.getSessionState(msg.Authorization) != sessionActive {
			t.Fatal("Unexpected session state")
		}
	})
}

func BenchmarkServerGenbody(b *testing.B) {
	handler := NewHandler("")
	for i := 0; i < b.N; i++ {
		count := maxSize
		handler.genbody(&count)
	}
}

func TestServerGenbody(t *testing.T) {
	t.Run("If size is too small", func(t *testing.T) {
		handler := NewHandler("")
		count := minSize - 100
		data, err := handler.genbody(&count)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != minSize {
			t.Fatal("Expected different size")
		}
	})

	t.Run("If size is too large", func(t *testing.T) {
		handler := NewHandler("")
		count := maxSize + 100
		data, err := handler.genbody(&count)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != maxSize {
			t.Fatal("Expected different size")
		}
	})
}

func TestServerDownload(t *testing.T) {
	t.Run("session missing", func(t *testing.T) {
		handler := NewHandler("")
		req := new(http.Request)
		w := httptest.NewRecorder()
		handler.download(w, req)
		resp := w.Result()
		if resp.StatusCode != 400 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("session expired", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		handler.maxIterations = 0
		req := new(http.Request)
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		w := httptest.NewRecorder()
		handler.download(w, req)
		resp := w.Result()
		if resp.StatusCode != 429 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("strcov.Atoi failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.URL = new(url.URL)
		req.URL.Path = "/dash/download/foobar"
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		w := httptest.NewRecorder()
		handler.download(w, req)
		resp := w.Result()
		if resp.StatusCode != 400 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("rand.Read failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		handler.deps.RandRead = func(p []byte) (n int, err error) {
			return 0, errors.New("Mocked error")
		}
		req := new(http.Request)
		req.URL = new(url.URL)
		req.URL.Path = "/dash/download"
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		w := httptest.NewRecorder()
		handler.download(w, req)
		resp := w.Result()
		if resp.StatusCode != 500 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("common case", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.URL = new(url.URL)
		req.URL.Path = "/dash/download/3500000"
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		w := httptest.NewRecorder()
		handler.download(w, req)
		resp := w.Result()
		if resp.StatusCode != 200 {
			t.Fatal("Expected different status code")
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 3500000 {
			t.Fatal("Expected different data length")
		}
	})
}

func TestServerSaveData(t *testing.T) {
	t.Run("os.MkdirAll failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		sessionInfo := handler.popSession(session)
		handler.deps.OSMkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("Mocked error")
		}
		err := handler.savedata(sessionInfo)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("os.OpenFile failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		sessionInfo := handler.popSession(session)
		handler.deps.OSMkdirAll = func(path string, perm os.FileMode) error {
			return nil
		}
		handler.deps.OSOpenFile = func(
			name string, flag int, perm os.FileMode,
		) (*os.File, error) {
			return nil, errors.New("Mocked error")
		}
		err := handler.savedata(sessionInfo)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("gzip.NewWriterLevel failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		sessionInfo := handler.popSession(session)
		handler.deps.OSMkdirAll = func(path string, perm os.FileMode) error {
			return nil
		}
		handler.deps.OSOpenFile = func(
			name string, flag int, perm os.FileMode,
		) (*os.File, error) {
			return ioutil.TempFile("", "neubot-dash-tests")
		}
		handler.deps.GzipNewWriterLevel = func(
			w io.Writer, level int,
		) (*gzip.Writer, error) {
			return nil, errors.New("Mocked error")
		}
		err := handler.savedata(sessionInfo)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Marshal failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		sessionInfo := handler.popSession(session)
		handler.deps.OSMkdirAll = func(path string, perm os.FileMode) error {
			return nil
		}
		handler.deps.OSOpenFile = func(
			name string, flag int, perm os.FileMode,
		) (*os.File, error) {
			return ioutil.TempFile("", "neubot-dash-tests")
		}
		handler.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		err := handler.savedata(sessionInfo)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("common case", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		sessionInfo := handler.popSession(session)
		handler.deps.OSMkdirAll = func(path string, perm os.FileMode) error {
			return nil
		}
		handler.deps.OSOpenFile = func(
			name string, flag int, perm os.FileMode,
		) (*os.File, error) {
			return ioutil.TempFile("", "neubot-dash-tests")
		}
		err := handler.savedata(sessionInfo)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestServerCollect(t *testing.T) {
	t.Run("session missing", func(t *testing.T) {
		handler := NewHandler("")
		req := new(http.Request)
		w := httptest.NewRecorder()
		handler.collect(w, req)
		resp := w.Result()
		if resp.StatusCode != 400 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		handler.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		w := httptest.NewRecorder()
		handler.collect(w, req)
		resp := w.Result()
		if resp.StatusCode != 400 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		handler.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return []byte("{"), nil
		}
		w := httptest.NewRecorder()
		handler.collect(w, req)
		resp := w.Result()
		if resp.StatusCode != 400 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("json.Marshal failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		handler.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return []byte("[]"), nil
		}
		handler.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		w := httptest.NewRecorder()
		handler.collect(w, req)
		resp := w.Result()
		if resp.StatusCode != 500 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("savedata failure", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		handler.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return []byte("[]"), nil
		}
		handler.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return []byte("[]"), nil
		}
		handler.deps.Savedata = func(session *sessionInfo) error {
			return errors.New("Mocked error")
		}
		w := httptest.NewRecorder()
		handler.collect(w, req)
		resp := w.Result()
		if resp.StatusCode != 500 {
			t.Fatal("Expected different status code")
		}
	})

	t.Run("common case", func(t *testing.T) {
		const session = "deadbeef"
		handler := NewHandler("")
		handler.createSession(session)
		req := new(http.Request)
		req.Header = make(http.Header)
		req.Header.Add(authorization, session)
		handler.deps.IOReadAll = func(r io.Reader) ([]byte, error) {
			return []byte("[]"), nil
		}
		handler.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return []byte("[]"), nil
		}
		handler.deps.Savedata = func(session *sessionInfo) error {
			return nil
		}
		w := httptest.NewRecorder()
		handler.collect(w, req)
		resp := w.Result()
		if resp.StatusCode != 200 {
			t.Fatal("Expected different status code")
		}
	})
}

func TestServerReaper(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	log.SetLevel(log.DebugLevel)
	handler := NewHandler("")
	handler.Logger = log.Log
	ctx, cancel := context.WithCancel(context.Background())
	handler.StartReaper(ctx)
	for i := 0; i < 17; i++ {
		handler.createSession(fmt.Sprintf("%d", i))
	}
	for handler.CountSessions() > 0 {
		time.Sleep(1 * time.Second)
	}
	cancel()
	handler.JoinReaper()
}
