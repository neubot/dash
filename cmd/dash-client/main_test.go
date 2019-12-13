package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/neubot/dash/client"
	"github.com/neubot/dash/server"
)

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	mux := http.NewServeMux()
	handler := server.NewHandler("../../testdata")
	ctx, cancel := context.WithCancel(context.Background())
	handler.Logger = log.Log
	handler.StartReaper(ctx)
	handler.RegisterHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	URL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	const parallel = 17
	errors := make([]error, 17)
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func(idx int) {
			time.Sleep(time.Duration(idx) * 100 * time.Millisecond)
			client := client.New(clientName, clientVersion)
			client.FQDN = URL.Host
			client.Scheme = "http" // we use httptest.NewServer
			errors[idx] = mainWithClientAndTimeout(client, 55*time.Second)
			wg.Done()
		}(i)
	}
	wg.Wait()
	cancel()
	handler.JoinReaper()
	for i := 0; i < parallel; i++ {
		if errors[i] != nil {
			t.Fatal(errors[i])
		}
	}
}
