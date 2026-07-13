package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/asanexample/alpha-shop/internal/flags"
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
	mux.HandleFunc("PATCH /api/cart/items/{productId}", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := session(w, r)
		if !ok {
			return
		}
		s.proxy(w, r, http.MethodPatch, s.cartURL+"/api/cart/"+url.PathEscape(sid)+"/items/"+url.PathEscape(r.PathValue("productId")), r.Body)
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
	// Both order-history routes require a signed-in account (an order has no stable identity to look up
	// by for a guest, and now carries a real address — orders' own ownership check needs a userId to
	// check against).
	mux.HandleFunc("GET /api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		u, err := s.currentUser(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sign in to view this order"})
			return
		}
		s.proxy(w, r, http.MethodGet, s.ordersURL+"/api/orders/"+url.PathEscape(r.PathValue("id"))+"?userId="+url.QueryEscape(u.UserID), nil)
	})
	mux.HandleFunc("GET /api/orders", func(w http.ResponseWriter, r *http.Request) {
		u, err := s.currentUser(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sign in to view your orders"})
			return
		}
		s.proxy(w, r, http.MethodGet, s.ordersURL+"/api/orders?userId="+url.QueryEscape(u.UserID), nil)
	})
}

// checkout proxies to the checkout service (ADR-057 — extracted so it's its own mutual-auth-secured hop,
// matching orders→payment). Requires an account: order history (GET /api/orders) has no stable identity
// to key on for a guest, so checkout is the point signup/login becomes mandatory rather than optional —
// this is also where a shopped-as-guest cart gets merged into the account. storefront still evaluates the
// checkout-experience flag (targetingKey = the user id, so a percentage rollout is sticky per account; the
// OpenFeature OTel hook stamps feature_flag.checkout-experience.* onto this request's span) and passes the
// resolved value through — checkout has no flags client of its own, it just orchestrates cart+orders with
// whatever it's told.
func (s *server) checkout(w http.ResponseWriter, r *http.Request) {
	u, err := s.currentUser(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sign in to complete your order"})
		return
	}
	var body struct {
		Card          string       `json:"card"`
		Address       checkoutAddr `json:"address"`
		PaymentMethod string       `json:"paymentMethod"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body) // optional fields (card/address are demo/best-effort)

	if anonSID := r.Header.Get(sessionHeader); anonSID != "" {
		s.mergeAnonymousCart(r.Context(), anonSID, u.UserID)
	}

	experience, _ := s.flags.StringValue(r.Context(), "checkout-experience", "standard", flags.EvalContext(u.UserID))
	reqBody, _ := json.Marshal(map[string]any{
		"sessionId":     u.UserID,
		"card":          body.Card,
		"experience":    experience,
		"address":       body.Address,
		"paymentMethod": body.PaymentMethod,
	})
	s.proxy(w, r, http.MethodPost, s.checkoutURL+"/api/checkout", bytes.NewReader(reqBody))
}

// checkoutAddr mirrors orders.Address (kept local for the same reason checkout's own copy is —
// storefront never inspects the fields, just passes the form's address through unchanged).
type checkoutAddr struct {
	Name    string `json:"name"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
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
