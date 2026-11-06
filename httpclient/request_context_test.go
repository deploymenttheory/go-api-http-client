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
