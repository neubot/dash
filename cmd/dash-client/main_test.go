package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/neubot/dash/client"
	"github.com/neubot/dash/server"
)

func init() {
	*flagY = true // acknowledge privacy policy for running integration tests
}

func TestRealmainSuccessful(t *testing.T) {
	testhelper(t, func(idx int, config testconfig) {
		time.Sleep(time.Duration(idx) * 100 * time.Millisecond)
		client := client.New(config.clientName, config.clientVersion)
		client.FQDN = config.fqdn
		client.Scheme = "http" // we use httptest.NewServer
		config.errors[idx] = realmain(config.ctx, client, 55*time.Second, nil)
	})
}

func TestCancelledContext(t *testing.T) {
	testhelper(t, func(idx int, config testconfig) {
		time.Sleep(time.Duration(idx) * 100 * time.Millisecond)
		client := client.New(config.clientName, config.clientVersion)
		client.FQDN = config.fqdn
		client.Scheme = "http" // we use httptest.NewServer
		ctx, cancel := context.WithCancel(config.ctx)
		cancel() // cause immediate failure
		err := realmain(ctx, client, 55*time.Second, nil)
		if !errors.Is(err, context.Canceled) {
			config.errors[idx] = fmt.Errorf("idx=%d: not the error we expected: %+w", idx, err)
		}
	})
}

func TestFailureBeforeEnd(t *testing.T) {
	testhelper(t, func(idx int, config testconfig) {
		time.Sleep(time.Duration(idx) * 100 * time.Millisecond)
		client := client.New(config.clientName, config.clientVersion)
		client.FQDN = config.fqdn
		client.Scheme = "http" // we use httptest.NewServer
		ctx, cancel := context.WithCancel(config.ctx)
		defer cancel()
		// note: the fourth argument causes cancel to be invoked after we
		// see the result of the first iteration
		err := realmain(ctx, client, 55*time.Second, cancel)
		if !errors.Is(err, context.Canceled) {
			config.errors[idx] = fmt.Errorf("idx=%d: not the error we expected: %+w", idx, err)
		}
	})
}

type testconfig struct {
	clientName    string
	clientVersion string
	ctx           context.Context // okay within same package
	errors        []error
	fqdn          string
}

func testhelper(t *testing.T, f func(int, testconfig)) {
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
			f(idx, testconfig{
				clientName:    clientName,
				clientVersion: clientVersion,
				ctx:           context.Background(),
				errors:        errors,
				fqdn:          URL.Host,
			})
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

func TestInternalMainCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately hang up
	internalmain(ctx)
}

func TestFmainSuccess(t *testing.T) {
	fmain(func(context.Context) error {
		return nil
	}, func(error, string, ...interface{}) {
		t.Fatal("should not be called")
	})
}

func TestFmainFailure(t *testing.T) {
	var called int32
	fmain(func(context.Context) error {
		return errors.New("antani")
	}, func(error, string, ...interface{}) {
		atomic.AddInt32(&called, 1)
	})
	if called != 1 {
		t.Fatal("not called")
	}
}

func TestMainOnly(t *testing.T) {
	mfunc := defaultMain
	defer func() {
		defaultMain = mfunc
	}()
	defaultMain = func(context.Context) error { return nil }
	main()
}
