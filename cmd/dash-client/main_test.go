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
	for i := 0; i < 17; i++ {
		wg.Add(1)
		go func(delay int) {
			time.Sleep(time.Duration(delay) * time.Second)
			client := client.New(clientName, clientVersion)
			client.FQDN = URL.Host
			err := mainWithClientAndTimeout(client, 55*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	cancel()
	handler.JoinReaper()
}
