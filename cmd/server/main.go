// Command server is app-alpha-shop: a generic starter service for the platform.
//
// It is deliberately minimal — a stdlib-only HTTP server exposing the liveness/readiness endpoint the
// platform deployment manifests probe (/healthz) and a JSON root handler. There is NO cloud/AWS
// dependency: an environment's AWS access (if any) is granted out-of-band via EKS Pod Identity to the named
// ServiceAccount (see k8s/preprod/serviceaccount.yaml) and declared in the Environment claim's `aws` block.
// Add an SDK + the access only when an app actually needs it.
//
// To start a NEW app from this template: copy the repo, rename app-alpha-shop -> app-<yourapp>, set your
// team/namespace/hostname in k8s/preprod/, and keep the thin .github/workflows callers as-is.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// httpClient makes the outbound call to alpha-checkout. A short timeout keeps a slow/hung downstream
// from tying up shop's request goroutines — never use the default (no-timeout) client for this.
var httpClient = &http.Client{Timeout: 3 * time.Second}

// newMux wires the routes — extracted so the unit test can exercise them without binding a port.
func newMux(version, namespace, checkoutURL string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Demonstrate a real east-west call: fetch a checkout confirmation and embed it, so a curl of
		// shop's public URL visibly shows checkout answered. Degrade gracefully — shop stays 200 with a
		// checkout.error field if the downstream is down, so shop's own readiness is unaffected.
		writeJSON(w, http.StatusOK, map[string]any{
			"app":       "app-alpha-shop",
			"version":   version,
			"namespace": namespace,
			"hostname":  r.Host,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"checkout":  callCheckout(r.Context(), checkoutURL),
		})
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return mux
}

func main() {
	version := getenv("VERSION", "dev")
	namespace := getenv("NAMESPACE", "unknown")
	// In-cluster Service DNS of alpha-checkout (this env's dev stage). Overridable per-stage.
	checkoutURL := getenv("CHECKOUT_URL", "http://app-alpha-checkout.alpha-checkout-dev.svc.cluster.local/checkout")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      newMux(version, namespace, checkoutURL),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Serve in the background; block on SIGTERM/SIGINT (k8s sends SIGTERM on pod termination), then drain
	// in-flight requests gracefully before exiting.
	go func() {
		log.Printf("starting app-alpha-shop version=%s namespace=%s", version, namespace)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down (draining in-flight requests)…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("stopped")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// callCheckout fetches an order confirmation from alpha-checkout. It never fails the caller: any error
// is returned as {"error": ...} so shop's own response stays 200 (the downstream being down must not
// flip shop's readiness). Returns the decoded checkout JSON on success.
func callCheckout(ctx context.Context, url string) any {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return map[string]string{"error": fmt.Sprintf("checkout returned %d", resp.StatusCode)}
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return map[string]string{"error": "decode checkout response: " + err.Error()}
	}
	return body
}
