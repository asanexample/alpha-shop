package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/asanexample/alpha-shop/internal/telemetry"
)

// sessionHeader carries the SPA-minted cart session id. It also becomes the payment/rollout targetingKey
// later (ADR-099) — a stable per-visitor key so a percentage rollout is sticky.
const sessionHeader = "X-Shop-Session"

// registerBuyPath wires the cart proxy + checkout orchestration onto the mux. The cart/orders services are
// internal; the BFF injects the session id (from the header) into their paths so the SPA never addresses them
// directly.
func (s *server) registerBuyPath(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cart", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := session(w, r)
		if !ok {
			return
		}
		s.proxy(w, r, http.MethodGet, s.cartURL+"/api/cart/"+url.PathEscape(sid), nil)
	})
	mux.HandleFunc("POST /api/cart/items", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := session(w, r)
		if !ok {
			return
		}
		s.proxy(w, r, http.MethodPost, s.cartURL+"/api/cart/"+url.PathEscape(sid)+"/items", r.Body)
	})
	mux.HandleFunc("DELETE /api/cart/items/{productId}", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := session(w, r)
		if !ok {
			return
		}
		s.proxy(w, r, http.MethodDelete, s.cartURL+"/api/cart/"+url.PathEscape(sid)+"/items/"+url.PathEscape(r.PathValue("productId")), nil)
	})
	mux.HandleFunc("DELETE /api/cart", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := session(w, r)
		if !ok {
			return
		}
		s.proxy(w, r, http.MethodDelete, s.cartURL+"/api/cart/"+url.PathEscape(sid), nil)
	})
	mux.HandleFunc("POST /api/checkout", s.checkout)
	mux.HandleFunc("GET /api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.proxy(w, r, http.MethodGet, s.ordersURL+"/api/orders/"+url.PathEscape(r.PathValue("id")), nil)
	})
}

// checkout builds an order from the session's cart, places it (orders→payment east-west), and clears the cart
// on a successful order — the whole flow one connected trace.
func (s *server) checkout(w http.ResponseWriter, r *http.Request) {
	sid, ok := session(w, r)
	if !ok {
		return
	}
	var body struct {
		Card string `json:"card"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body) // optional (card is a demo field)

	cart, err := s.fetchCart(r.Context(), sid)
	if err != nil {
		s.fail(w, r, "checkout: fetch cart", err)
		return
	}
	if len(cart.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "your cart is empty"})
		return
	}

	orderReq := map[string]any{"sessionId": sid, "lines": cart.Items, "card": body.Card}
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
		if _, _, err := s.call(r.Context(), http.MethodDelete, s.cartURL+"/api/cart/"+url.PathEscape(sid), nil); err != nil {
			telemetry.Logger.ErrorContext(r.Context(), "checkout: clear cart failed (continuing)", "err", err)
		}
	}
	telemetry.Logger.InfoContext(r.Context(), "checkout", "order", order.ID, "status", order.Status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

// cartLine mirrors a cart line (kept local so the BFF doesn't pull the AWS-heavy cart package).
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

// proxy forwards a request to an internal service and streams the response back verbatim.
func (s *server) proxy(w http.ResponseWriter, r *http.Request, method, upstream string, body io.Reader) {
	raw, status, err := s.call(r.Context(), method, upstream, body)
	if err != nil {
		s.fail(w, r, "upstream", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
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

func session(w http.ResponseWriter, r *http.Request) (string, bool) {
	sid := r.Header.Get(sessionHeader)
	if sid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing " + sessionHeader + " header"})
		return "", false
	}
	return sid, true
}
