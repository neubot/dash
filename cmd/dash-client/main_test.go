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
	"github.com/neubot/dash/server"
)

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	mux := http.NewServeMux()
	handler := server.NewHandler("../../testdata")
	ctx, cancel := context.WithCancel(context.Background())
	handler.StartReaper(ctx)
	handler.RegisterHandlers(mux)
	handler.Logger = log.Log
	server := httptest.NewServer(mux)
	defer server.Close()
	URL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	*flagHostname = URL.Host
	var wg sync.WaitGroup
	for i := 0; i < 17; i++ {
		wg.Add(1)
		go func(delay int) {
			time.Sleep(time.Duration(delay) * time.Second)
			err = internalmain()
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
