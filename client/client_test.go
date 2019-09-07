package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	dash "github.com/neubot/dash/client"
	"github.com/neubot/dash/internal/mocks"
)

const (
	softwareName    = "dash-client-go-test"
	softwareVersion = "0.0.1"
)

const (
	expectedMlabNSURL  = "https://locate.measurementlab.net/neubot"
	expectedServerFQDN = "cdn.neubot.org"
)

var (
	expectedNegotiateURL = fmt.Sprintf("http://%s/negotiate/dash", expectedServerFQDN)

	expectedFirstDownloadURL = fmt.Sprintf("http://%s/dash/download/750000", expectedServerFQDN)

	expectedCollectURL = fmt.Sprintf("http://%s/collect/dash", expectedServerFQDN)
)

func mlabnsSuccessfulResponse() mocks.HTTPRoundTripInfo {
	return mocks.HTTPRoundTripInfo{
		Error: nil,
		URL:   expectedMlabNSURL,
		Response: &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(strings.NewReader(
				fmt.Sprintf(`{"fqdn": "%s"}`, expectedServerFQDN),
			)),
		},
	}
}

func responseWith404(URL string) mocks.HTTPRoundTripInfo {
	return mocks.HTTPRoundTripInfo{
		Error: nil,
		URL:   URL,
		Response: &http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(strings.NewReader("Not Found")),
		},
	}
}

func responseWithReadBodyError(URL string) mocks.HTTPRoundTripInfo {
	return mocks.HTTPRoundTripInfo{
		Error: nil,
		URL:   URL,
		Response: &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(mocks.ProgrammableReader{
				Err: mocks.ErrMocked,
			}),
		},
	}
}

func responseWithInvalidJSON(URL string) mocks.HTTPRoundTripInfo {
	return mocks.HTTPRoundTripInfo{
		Error: nil,
		URL:   URL,
		Response: &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(mocks.ProgrammableReader{
				Reader: bytes.NewReader([]byte("{")),
			}),
		},
	}
}

func responseWithEmptyJSON(URL string) mocks.HTTPRoundTripInfo {
	return mocks.HTTPRoundTripInfo{
		Error: nil,
		URL:   URL,
		Response: &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(mocks.ProgrammableReader{
				Reader: bytes.NewReader([]byte("{}")),
			}),
		},
	}
}

func goodResponse(URL, body string) mocks.HTTPRoundTripInfo {
	return mocks.HTTPRoundTripInfo{
		Error: nil,
		URL:   URL,
		Response: &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(mocks.ProgrammableReader{
				Reader: strings.NewReader(body),
			}),
		},
	}
}

func goodNegotiationResponse() mocks.HTTPRoundTripInfo {
	return goodResponse(expectedNegotiateURL, fmt.Sprintf(`{
		"authorization": "DEADBEEF",
		"unchoked": 1
	}`))
}

func goodCollectResponse() mocks.HTTPRoundTripInfo {
	return goodResponse(expectedCollectURL, fmt.Sprintf(`[{
		"iteration": 0,
		"ticks": 1.1,
		"timestamp": 123456789
	}]`))
}

func TestMlabNSFailure(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mocks.NewHTTPRoundTripFailure(expectedMlabNSURL),
	)
	_, err := clnt.StartDownload(context.Background())
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiateMarshalJSONError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.Dependencies.JSONMarshal = func(v interface{}) ([]byte, error) {
		return nil, mocks.ErrMocked
	}
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiateHTTPNewRequestError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.Dependencies.HTTPNewRequest = func(method string, url string, body io.Reader) (*http.Request, error) {
		return nil, mocks.ErrMocked
	}
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiateRequestError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		mocks.NewHTTPRoundTripFailure(expectedNegotiateURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiate404Error(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		responseWith404(expectedNegotiateURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiateReadBodyError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		responseWithReadBodyError(expectedNegotiateURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiateJSONParseError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		responseWithInvalidJSON(expectedNegotiateURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestNegotiateNotAuthorized(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		responseWithEmptyJSON(expectedNegotiateURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestDownloadHTTPNewRequestError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		mocks.NewHTTPRoundTripFailure(expectedFirstDownloadURL),
	)
	var calls int
	clnt.Dependencies.HTTPNewRequest = func(method string, url string, body io.Reader) (*http.Request, error) {
		if calls <= 0 {
			calls++
			return http.NewRequest(method, url, body)
		}
		return nil, mocks.ErrMocked
	}
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestDownloadRequestError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		mocks.NewHTTPRoundTripFailure(expectedFirstDownloadURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestDownloadRequest404(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		responseWith404(expectedFirstDownloadURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestDownloadReadBodyError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		responseWithReadBodyError(expectedFirstDownloadURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for range ch {
		t.Fatal("Did not expect a meaurement here")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestCollectMarshalJSONError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		mocks.NewHTTPRoundTripFailure(expectedCollectURL),
	)
	var calls int
	clnt.Dependencies.JSONMarshal = func(v interface{}) ([]byte, error) {
		if calls <= 0 {
			calls++
			return json.Marshal(v)
		}
		return nil, mocks.ErrMocked
	}
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestCollectHTTPNewRequest(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		mocks.NewHTTPRoundTripFailure(expectedCollectURL),
	)
	var calls int
	clnt.Dependencies.HTTPNewRequest = func(method string, url string, body io.Reader) (*http.Request, error) {
		if calls <= 1 {
			calls++
			return http.NewRequest(method, url, body)
		}
		return nil, mocks.ErrMocked
	}
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestCollectRequestError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		mocks.NewHTTPRoundTripFailure(expectedCollectURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestCollectRequest404(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		responseWith404(expectedCollectURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestCollectReadBodyError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		responseWithReadBodyError(expectedCollectURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestCollectInvalidJSONError(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		responseWithInvalidJSON(expectedCollectURL),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err == nil {
		t.Fatal("Expected an error here")
	}
}

func TestAllGood(t *testing.T) {
	clnt := dash.NewClient(softwareName, softwareVersion)
	clnt.NumIterations = 1
	clnt.MLabNSClient.HTTPClient = mocks.NewHTTPClient(
		mlabnsSuccessfulResponse(),
	)
	clnt.HTTPClient = mocks.NewHTTPClient(
		goodNegotiationResponse(),
		goodResponse(expectedFirstDownloadURL, "VERYSHORTBODY"),
		goodCollectResponse(),
	)
	ch, err := clnt.StartDownload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for range ch {
		count++
	}
	if count != 1 {
		t.Fatal("Expected to see a single measurement")
	}
	err = clnt.Error()
	if err != nil {
		t.Fatal(err)
	}
	if len(clnt.ServerResults()) <= 0 {
		t.Fatal("Missing server results")
	}
}
