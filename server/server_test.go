package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/neubot/dash/common"
	"github.com/neubot/dash/internal/mocks"
	dash "github.com/neubot/dash/server"
)

func prepareEx() (mux *http.ServeMux, handler *dash.Handler) {
	mux = http.NewServeMux()
	handler = dash.NewHandler("..") // run in toplevel dir
	handler.RegisterHandlers(mux)
	return
}

func prepare() (mux *http.ServeMux) {
	mux, _ = prepareEx()
	return
}

func TestNegotiateNoRemoteAddr(t *testing.T) {
	mux := prepare()
	req := httptest.NewRequest("POST", "/negotiate/dash", nil)
	req.RemoteAddr = ""
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 500 {
		t.Fatal("Expected different status code")
	}
}

func doNegotiate(mux *http.ServeMux) (common.NegotiateResponse, error) {
	var negotiateResponse common.NegotiateResponse
	req := httptest.NewRequest("POST", "/negotiate/dash", nil)
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 200 {
		return negotiateResponse, errors.New("Invalid status code")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return negotiateResponse, err
	}
	err = json.Unmarshal(data, &negotiateResponse)
	if err != nil {
		return negotiateResponse, err
	}
	return negotiateResponse, nil
}

func TestNegotiateNormal(t *testing.T) {
	mux := prepare()
	negotiateResponse, err := doNegotiate(mux)
	if err != nil {
		t.Fatal(err)
	}
	if len(negotiateResponse.Authorization) <= 0 {
		t.Fatal("Authorization is empty")
	}
	if negotiateResponse.QueuePos != 0 {
		t.Fatal("Unexpected queue position")
	}
	if net.ParseIP(negotiateResponse.RealAddress) == nil {
		t.Fatal("Cannot parse RealAddress")
	}
	if negotiateResponse.Unchoked != 1 {
		t.Fatal("Unexpected unchoked value")
	}
}

func TestDownloadNoAuth(t *testing.T) {
	mux := prepare()
	req := httptest.NewRequest("GET", "/dash/download", nil)
	req.RemoteAddr = ""
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 400 {
		t.Fatal("Expected different status code")
	}
}

func TestDownloadInvalidSize(t *testing.T) {
	mux := prepare()
	negotiateResponse, err := doNegotiate(mux)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/dash/download/xyz", nil)
	req.Header.Set("Authorization", negotiateResponse.Authorization)
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 400 {
		t.Fatal("Expected different status code")
	}
}

func doDownloadWithAuth(mux *http.ServeMux, urlpath, auth string) *http.Response {
	req := httptest.NewRequest("GET", urlpath, nil)
	req.Header.Set("Authorization", auth)
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	return writer.Result()
}

func doDownload(mux *http.ServeMux, urlpath string) (string, int, error) {
	negotiateResponse, err := doNegotiate(mux)
	if err != nil {
		return "", 0, err
	}
	resp := doDownloadWithAuth(mux, urlpath, negotiateResponse.Authorization)
	if resp.StatusCode != 200 {
		return "", 0, errors.New("Expected different status code")
	}
	if resp.Header.Get("Content-Type") != "video/mp4" {
		return "", 0, errors.New("Unexpected Content-Type value")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	return negotiateResponse.Authorization, len(data), nil
}

func TestDownloadNoSize(t *testing.T) {
	mux := prepare()
	_, size, err := doDownload(mux, "/dash/download")
	if err != nil {
		t.Fatal(err)
	}
	if size != dash.MinSize {
		t.Fatal("Unexpected segment length")
	}
}

func TestDownloadNegativeSize(t *testing.T) {
	mux := prepare()
	_, size, err := doDownload(mux, "/dash/download/-1")
	if err != nil {
		t.Fatal(err)
	}
	if size != dash.MinSize {
		t.Fatal("Unexpected segment length")
	}
}

func TestDownloadHugeSize(t *testing.T) {
	mux := prepare()
	_, size, err := doDownload(mux, "/dash/download/1125899906842624")
	if err != nil {
		t.Fatal(err)
	}
	if size != dash.MaxSize {
		t.Fatal("Unexpected segment length")
	}
}

func TestDownloadTooManyRequests(t *testing.T) {
	mux, handler := prepareEx()
	handler.MaxIterations = 1
	auth, _, err := doDownload(mux, "/dash/download")
	if err != nil {
		t.Fatal(err)
	}
	resp := doDownloadWithAuth(mux, "/dash/download", auth)
	if resp.StatusCode != 429 {
		t.Fatal("Unexpected status code")
	}
}

func TestCollectNoAuth(t *testing.T) {
	mux := prepare()
	req := httptest.NewRequest("POST", "/collect/dash", nil)
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 400 {
		t.Fatal("Expected different status code")
	}
}

func TestCollectReadBodyError(t *testing.T) {
	mux := prepare()
	negotiateResponse, err := doNegotiate(mux)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/collect/dash", nil)
	req.Header.Set("Authorization", negotiateResponse.Authorization)
	req.Body = ioutil.NopCloser(mocks.ProgrammableReader{
		Err: mocks.ErrMocked,
	})
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 400 {
		t.Fatal("Expected different status code")
	}
}

func TestCollectNoBody(t *testing.T) {
	mux := prepare()
	negotiateResponse, err := doNegotiate(mux)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/collect/dash", nil)
	req.Header.Set("Authorization", negotiateResponse.Authorization)
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 400 {
		t.Fatal("Expected different status code")
	}
}

func TestCollectGood(t *testing.T) {
	mux := prepare()
	negotiateResponse, err := doNegotiate(mux)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/collect/dash", nil)
	req.Header.Set("Authorization", negotiateResponse.Authorization)
	req.Body = ioutil.NopCloser(mocks.ProgrammableReader{
		Reader: strings.NewReader("[]"),
	})
	writer := httptest.NewRecorder()
	mux.ServeHTTP(writer, req)
	resp := writer.Result()
	if resp.StatusCode != 200 {
		t.Fatal("Expected different status code")
	}
}

func TestReaper(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	log.SetLevel(log.DebugLevel)
	mux, handler := prepareEx()
	handler.Logger = log.Log
	ctx, cancel := context.WithCancel(context.Background())
	handler.StartReaper(ctx)
	for i := 0; i < 17; i++ {
		_, err := doNegotiate(mux)
		if err != nil {
			t.Fatal(err)
		}
	}
	for handler.CountSessions() > 0 {
		time.Sleep(1 * time.Second)
	}
	cancel()
	handler.JoinReaper()
}
