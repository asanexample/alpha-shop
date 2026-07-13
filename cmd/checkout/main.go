// Command checkout is the bike-shop checkout service (team alpha, product shop, service checkout).
//
// Internal, east-west only (ClusterIP, no HTTPRoute): storefront's BFF proxies POST /api/checkout here.
// Orchestrates the buy path — fetch the session's cart, place the order (east-west call to orders, which in
// turn authorizes payment and may kick off a Bravo Dispatch shipment), then clear the cart on a successful
// order — one connected trace: storefront -> checkout -> cart / orders. Extracted from storefront's
// buypath.go (ADR-057) so checkout is its own mutual-auth-secured hop, matching orders -> payment. Storefront
// still evaluates the checkout-experience feature flag (it already has the flagship sync wiring) and passes
// the resolved value through — checkout has no flags client of its own.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asanexample/alpha-shop/internal/telemetry"
)

type server struct {
	httpc     *http.Client
	cartURL   string
	ordersURL string
}

// cartLine mirrors a cart line (kept local so checkout doesn't pull the AWS-heavy cart package).
type cartLine struct {
	ProductID  string `json:"productId"`
	Slug       string `json:"slug,omitempty"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Qty        int    `json:"qty"`
}

type cartView struct {
	Items []cartLine `json:"items"`
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/checkout", s.checkout)

	return mux
}

// checkout builds an order from the session's cart, places it (checkout -> orders east-west, which in turn
// authorizes payment), and clears the cart on a successful order.
func (s *server) checkout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID  string `json:"sessionId"`
		Card       string `json:"card"`
		Experience string `json:"experience"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing sessionId"})
		return
	}

	cart, err := s.fetchCart(r.Context(), body.SessionID)
	if err != nil {
		s.fail(w, r, "checkout: fetch cart", err)
		return
	}
	if len(cart.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "your cart is empty"})
		return
	}

	orderReq := map[string]any{"sessionId": body.SessionID, "lines": cart.Items, "card": body.Card, "experience": body.Experience}
	reqBody, _ := json.Marshal(orderReq)
	raw, status, err := s.call(r.Context(), http.MethodPost, s.ordersURL+"/api/orders", bytes.NewReader(reqBody))
	if err != nil {
		s.fail(w, r, "checkout: place order", err)
		return
	}
	if status != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(raw)
		return
	}

	var order struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(raw, &order)
	if order.Status == "placed" {
		// Empty the cart now the order is paid (best-effort; the order is already durable).
		if _, _, err := s.call(r.Context(), http.MethodDelete, s.cartURL+"/api/cart/"+url.PathEscape(body.SessionID), nil); err != nil {
			telemetry.Logger.ErrorContext(r.Context(), "checkout: clear cart failed (continuing)", "err", err)
		}
	}
	telemetry.Logger.InfoContext(r.Context(), "checkout", "order", order.ID, "status", order.Status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

// fetchCart reads the session's cart from the cart service (the {cart:{items}} envelope).
func (s *server) fetchCart(ctx context.Context, sid string) (cartView, error) {
	raw, status, err := s.call(ctx, http.MethodGet, s.cartURL+"/api/cart/"+url.PathEscape(sid), nil)
	if err != nil {
		return cartView{}, err
	}
	if status != http.StatusOK {
		return cartView{}, io.ErrUnexpectedEOF
	}
	var env struct {
		Cart cartView `json:"cart"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return cartView{}, err
	}
	return env.Cart, nil
}

// call does a trace-propagating request to an internal service and returns the body + status.
func (s *server) call(ctx context.Context, method, upstream string, body io.Reader) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, upstream, body)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.httpc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return raw, resp.StatusCode, err
}

func (s *server) fail(w http.ResponseWriter, r *http.Request, what string, err error) {
	telemetry.Logger.ErrorContext(r.Context(), "checkout upstream failed", "what", what, "err", err)
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream unavailable: " + what})
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-checkout")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	srv := &server{
		httpc:     telemetry.Client(),
		cartURL:   getenv("CART_URL", "http://cart"),
		ordersURL: getenv("ORDERS_URL", "http://orders"),
	}
	handler := telemetry.WrapHandler(srv.routes(), "http.server")

	httpSrv := &http.Server{Addr: getenv("ADDR", ":8080"), Handler: handler, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-checkout", "addr", httpSrv.Addr)
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
