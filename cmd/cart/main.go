// Command cart is the bike-shop cart service (team alpha, product shop, service cart).
//
// Internal, east-west only (ClusterIP, no HTTPRoute): the storefront BFF calls it to manage a session's
// cart. Cart state persists in the self-service DynamoDB table published as ITEMS_TABLE in the cart-resources
// ConfigMap (ADR-073); absent → in-memory (local dev). Instrumented via internal/telemetry.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
	"github.com/asanexample/alpha-shop/internal/cart"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

func newMux(store *cart.Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/cart/{sessionId}", func(w http.ResponseWriter, r *http.Request) {
		c, err := store.Get(r.Context(), r.PathValue("sessionId"))
		if err != nil {
			fail(w, r, "get cart", err)
			return
		}
		writeCart(w, c)
	})

	mux.HandleFunc("POST /api/cart/{sessionId}/items", func(w http.ResponseWriter, r *http.Request) {
		var item cart.Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}
		if item.ProductID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "productId is required"})
			return
		}
		c, err := store.Add(r.Context(), r.PathValue("sessionId"), item)
		if err != nil {
			fail(w, r, "add item", err)
			return
		}
		telemetry.Logger.InfoContext(r.Context(), "cart add", "product", item.ProductID, "count", c.Count())
		writeCart(w, c)
	})

	mux.HandleFunc("DELETE /api/cart/{sessionId}/items/{productId}", func(w http.ResponseWriter, r *http.Request) {
		c, err := store.Remove(r.Context(), r.PathValue("sessionId"), r.PathValue("productId"))
		if err != nil {
			fail(w, r, "remove item", err)
			return
		}
		writeCart(w, c)
	})

	mux.HandleFunc("PATCH /api/cart/{sessionId}/items/{productId}", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Qty int `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}
		c, err := store.SetQty(r.Context(), r.PathValue("sessionId"), r.PathValue("productId"), body.Qty)
		if err != nil {
			fail(w, r, "set quantity", err)
			return
		}
		writeCart(w, c)
	})

	mux.HandleFunc("DELETE /api/cart/{sessionId}", func(w http.ResponseWriter, r *http.Request) {
		if err := store.Clear(r.Context(), r.PathValue("sessionId")); err != nil {
			fail(w, r, "clear cart", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-cart")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	kv, err := awskv.Open(ctx, os.Getenv("ITEMS_TABLE"))
	if err != nil {
		telemetry.Logger.Error("kv init failed", "err", err)
		os.Exit(1)
	}
	store := cart.New(kv)
	telemetry.Logger.Info("cart store", "backend", store.Backend())

	srv := &http.Server{Addr: getenv("ADDR", ":8080"), Handler: telemetry.WrapHandler(newMux(store), "http.server"), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-cart", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			telemetry.Logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	telemetry.Logger.Info("shutting down (draining in-flight requests)…")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		telemetry.Logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	telemetry.Logger.Info("stopped")
}

// writeCart returns the cart plus derived totals the UI needs.
func writeCart(w http.ResponseWriter, c cart.Cart) {
	writeJSON(w, http.StatusOK, map[string]any{
		"cart":          c,
		"count":         c.Count(),
		"subtotalCents": c.SubtotalCents(),
	})
}

func fail(w http.ResponseWriter, r *http.Request, what string, err error) {
	telemetry.Logger.ErrorContext(r.Context(), "cart error", "what", what, "err", err)
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": what + ": " + err.Error()})
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
