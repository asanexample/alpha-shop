package flags

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	flagdsync "github.com/open-feature/flagd/core/pkg/sync"
)

// newSyncer points an httpSyncer at a test server serving body/status.
func newSyncer(url string) *httpSyncer {
	return &httpSyncer{url: url, interval: time.Hour, client: &http.Client{Timeout: time.Second}}
}

// drain returns the DataSync waiting on out, or ok=false if none is buffered.
func drain(out chan flagdsync.DataSync) (flagdsync.DataSync, bool) {
	select {
	case d := <-out:
		return d, true
	default:
		return flagdsync.DataSync{}, false
	}
}

// emit pushes the flag-set once, dedups an unchanged set, and re-emits on change.
func TestEmit_DedupAndChange(t *testing.T) {
	body := `{"flags":{"checkout-experience":{"state":"ENABLED","defaultVariant":"standard"}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	s := newSyncer(srv.URL)
	out := make(chan flagdsync.DataSync, 1)

	s.emit(context.Background(), out)
	d, ok := drain(out)
	if !ok {
		t.Fatal("first emit: expected a DataSync snapshot")
	}
	if d.FlagData != body {
		t.Fatalf("first emit: flag data passed through unchanged: got %q", d.FlagData)
	}
	if !s.IsReady() {
		t.Fatal("first emit: syncer should be ready after a successful fetch")
	}

	// Identical set → no re-emit (dedup).
	s.emit(context.Background(), out)
	if _, ok := drain(out); ok {
		t.Fatal("second emit: unchanged set must not re-emit")
	}
}

// A fetch error must NOT emit and must NOT flip ready — the resolver keeps its last-good set (fail-static).
func TestEmit_FailStatic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	url := srv.URL
	srv.Close() // force a connection error, not just a 5xx

	s := newSyncer(url)
	out := make(chan flagdsync.DataSync, 1)
	s.emit(context.Background(), out)

	if _, ok := drain(out); ok {
		t.Fatal("fetch error must not emit a snapshot")
	}
	if s.IsReady() {
		t.Fatal("fetch error must not flip the syncer ready")
	}
}

// ReSync clears the dedup memory so the next emit re-pushes even an unchanged set.
func TestReSync_ForcesReEmit(t *testing.T) {
	body := `{"flags":{}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	s := newSyncer(srv.URL)
	out := make(chan flagdsync.DataSync, 2)

	s.emit(context.Background(), out)
	if _, ok := drain(out); !ok {
		t.Fatal("initial emit expected")
	}
	if err := s.ReSync(context.Background(), out); err != nil {
		t.Fatalf("ReSync error: %v", err)
	}
	if _, ok := drain(out); !ok {
		t.Fatal("ReSync must re-emit the current set even though it is unchanged")
	}
}
