package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoRequestWithContextCancellation(t *testing.T) {
	for _, retries := range []bool{false, true} {
		t.Run(map[bool]string{false: "single request", true: "retrying request"}[retries], func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				calls.Add(1)
				<-req.Context().Done()
			}))
			defer server.Close()
			client := newTestClient(server.URL)
			client.config.RetryEligiableRequests = retries
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			resp, err := client.DoRequestWithContext(ctx, http.MethodGet, "/blocked", nil, nil)
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expected deadline error, got %v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("expected one request, got %d", calls.Load())
			}
		})
	}
}

func TestDoRequestWithContextCancelsBackoff(t *testing.T) {
	var calls int32
	server := serverAlways500(t, &calls)
	client := newTestClient(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	resp, err := client.DoRequestWithContext(ctx, http.MethodGet, "/retry", nil, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error during backoff, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatal("cancelled backoff must not issue another request")
	}
}

func TestDoRequestWithContextDeadlineReplacesDefaultTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		select {
		case <-time.After(60 * time.Millisecond):
			w.WriteHeader(http.StatusNoContent)
		case <-req.Context().Done():
		}
	}))
	defer server.Close()
	client := newTestClient(server.URL)
	// Scale the implicit ten-second timeout down to keep the regression fast.
	client.http.Timeout = 20 * time.Millisecond
	client.config.Timeout = 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.DoRequestWithContext(ctx, http.MethodPut, "/slow-write", nil, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("operation deadline must allow a write longer than the implicit default: %v", err)
	}
	if client.http.Timeout != 20*time.Millisecond {
		t.Fatal("context request changed the shared HTTP client timeout")
	}
}

func TestDoRequestWithContextPreservesExplicitTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		<-req.Context().Done()
	}))
	defer server.Close()
	client := newTestClient(server.URL)
	client.http.Timeout = 20 * time.Millisecond
	client.config.Timeout = 20 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.DoRequestWithContext(ctx, http.MethodPut, "/slow-write", nil, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if !errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		t.Fatalf("explicit client timeout should expire before the operation context: %v", err)
	}
}

func TestDoRequestWithoutDeadlinePreservesDefaultTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		<-req.Context().Done()
	}))
	defer server.Close()
	client := newTestClient(server.URL)
	client.http.Timeout = 20 * time.Millisecond
	client.config.Timeout = 0
	resp, err := client.DoRequest(http.MethodGet, "/slow-read", nil, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("legacy requests must retain their default timeout: %v", err)
	}
}
