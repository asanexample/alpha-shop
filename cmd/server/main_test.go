package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthz asserts the liveness/readiness endpoint the platform probes returns 200 {"status":"ok"}.
func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(newMux("test", "test-ns", "http://checkout.invalid/checkout"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want \"ok\"", body["status"])
	}
}

// TestRootEmbedsCheckout asserts GET / makes the downstream call and embeds checkout's response.
func TestRootEmbedsCheckout(t *testing.T) {
	checkout := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order":"0001","confirmedBy":"app-alpha-checkout","host":"stub"}`))
	}))
	defer checkout.Close()

	srv := httptest.NewServer(newMux("test", "test-ns", checkout.URL+"/checkout"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	co, ok := body["checkout"].(map[string]any)
	if !ok {
		t.Fatalf("checkout field missing or wrong type: %#v", body["checkout"])
	}
	if co["confirmedBy"] != "app-alpha-checkout" {
		t.Fatalf("checkout.confirmedBy = %v, want \"app-alpha-checkout\"", co["confirmedBy"])
	}
}
