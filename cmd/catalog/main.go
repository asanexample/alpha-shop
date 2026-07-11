// Command catalog is the bike-shop catalog service (team alpha, product shop, service catalog).
//
// It is an internal, east-west-only service (ClusterIP, no HTTPRoute): the storefront BFF calls it to render
// the browse pages. It serves a read-only JSON API over the seeded catalog and is instrumented out of the box
// (OTel traces + Pyroscope profiles via internal/telemetry), so a storefront→catalog call shows up as one
// connected distributed trace with per-service spans.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/asanexample/alpha-shop/internal/catalog"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

func newMux(store *catalog.Store) *http.ServeMux {
	mux := http.NewServeMux()
	log := telemetry.Logger

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/catalog/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.Categories())
	})

	mux.HandleFunc("GET /api/catalog/brands", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.Brands())
	})

	mux.HandleFunc("GET /api/catalog/products", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f := catalog.Filter{
			Category:      q.Get("category"),
			Kind:          catalog.Kind(q.Get("kind")),
			Brand:         q.Get("brand"),
			Query:         q.Get("q"),
			MinPriceCents: atoiDollars(q.Get("minPrice")),
			MaxPriceCents: atoiDollars(q.Get("maxPrice")),
			OnSaleOnly:    q.Get("onSale") == "true",
			FeaturedOnly:  q.Get("featured") == "true",
		}
		products := store.List(f)
		log.InfoContext(r.Context(), "list products", "count", len(products), "category", f.Category, "brand", f.Brand, "q", f.Query)
		writeJSON(w, http.StatusOK, map[string]any{"products": products, "count": len(products)})
	})

	mux.HandleFunc("GET /api/catalog/products/{idOrSlug}", func(w http.ResponseWriter, r *http.Request) {
		p, ok := store.Product(r.PathValue("idOrSlug"))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"product": p, "related": store.Related(p, 4)})
	})

	return mux
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-catalog")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	store := catalog.New()
	handler := telemetry.WrapHandler(newMux(store), "http.server")

	srv := &http.Server{Addr: ":8080", Handler: handler, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-catalog", "addr", srv.Addr)
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

// atoiDollars parses a whole-dollar query value into cents (e.g. "1500" → 150000). Returns 0 on empty/invalid.
func atoiDollars(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n * 100
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
