// Command storefront is the bike-shop storefront (team alpha, product shop, service storefront).
//
// It is the shop's front door: the ONLY edge-exposed service (its HTTPRoute serves the public host). It does
// two jobs in one Kyverno-compliant image (the flagship pattern): it serves the embedded React SPA, and it
// runs a BFF API the SPA calls, which aggregates the internal services (catalog today; cart/orders next) over
// east-west calls. Those calls propagate the trace (internal/telemetry), so a page load is one connected trace
// spanning storefront→catalog.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asanexample/alpha-shop/internal/catalogclient"
	"github.com/asanexample/alpha-shop/internal/telemetry"
	"github.com/asanexample/alpha-shop/web"
)

type server struct {
	catalog *catalogclient.Client
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	log := telemetry.Logger

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Nav data for the header mega-menu + facet UI: all categories + brands. Loaded once by the SPA.
	mux.HandleFunc("GET /api/nav", func(w http.ResponseWriter, r *http.Request) {
		cats, err := s.catalog.Categories(r.Context())
		if err != nil {
			s.fail(w, r, "nav categories", err)
			return
		}
		brands, err := s.catalog.Brands(r.Context())
		if err != nil {
			s.fail(w, r, "nav brands", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"categories": cats, "brands": brands})
	})

	// Homepage: featured products + the category list (for hero tiles).
	mux.HandleFunc("GET /api/home", func(w http.ResponseWriter, r *http.Request) {
		featured, err := s.catalog.Products(r.Context(), url.Values{"featured": {"true"}})
		if err != nil {
			s.fail(w, r, "home featured", err)
			return
		}
		cats, err := s.catalog.Categories(r.Context())
		if err != nil {
			s.fail(w, r, "home categories", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"featured": featured, "categories": cats})
	})

	// Product listing / search — proxies the catalog filter query straight through.
	mux.HandleFunc("GET /api/products", func(w http.ResponseWriter, r *http.Request) {
		products, err := s.catalog.Products(r.Context(), r.URL.Query())
		if err != nil {
			s.fail(w, r, "products", err)
			return
		}
		log.InfoContext(r.Context(), "storefront listing", "count", len(products))
		writeJSON(w, http.StatusOK, map[string]any{"products": products, "count": len(products)})
	})

	// Product detail + related.
	mux.HandleFunc("GET /api/products/{idOrSlug}", func(w http.ResponseWriter, r *http.Request) {
		detail, ok, err := s.catalog.Product(r.Context(), r.PathValue("idOrSlug"))
		if err != nil {
			s.fail(w, r, "product detail", err)
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
			return
		}
		writeJSON(w, http.StatusOK, detail)
	})

	// The SPA (embedded under -tags storefront). "/" is the catch-all; the /api and /healthz patterns above
	// are more specific, so they win. Absent (plain build) → API only.
	if h := web.Handler(); h != nil {
		mux.Handle("/", h)
		log.Info("storefront UI: embedded")
	} else {
		log.Info("storefront UI: not embedded (build with -tags storefront)")
	}

	return mux
}

func (s *server) fail(w http.ResponseWriter, r *http.Request, what string, err error) {
	// A failed east-west call must not blank the page hard; log it (with trace_id) and return a clean 502.
	telemetry.Logger.ErrorContext(r.Context(), "bff upstream failed", "what", what, "err", err)
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream unavailable: " + what})
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-storefront")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	catalogURL := getenv("CATALOG_URL", "http://catalog")
	srv := &server{catalog: catalogclient.New(catalogURL)}
	handler := telemetry.WrapHandler(srv.routes(), "http.server")

	httpSrv := &http.Server{Addr: getenv("ADDR", ":8080"), Handler: handler, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-storefront", "addr", httpSrv.Addr, "catalog", catalogURL)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	if err := httpSrv.Shutdown(ctxShutdown); err != nil {
		telemetry.Logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	telemetry.Logger.Info("stopped")
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
