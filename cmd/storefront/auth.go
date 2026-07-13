package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/asanexample/alpha-shop/internal/telemetry"
)

// sessionCookie carries the accounts-issued opaque token. HttpOnly (never readable by JS — the token
// is a bearer credential); SameSite=Lax deliberately (blocks the cookie riding along on cross-site
// POSTs, giving basic CSRF protection on /api/checkout for free, no separate CSRF token needed for a
// demo); Secure (HTTPS only — every real deployment of this app is behind the Gateway's TLS).
const sessionCookie = "shop_session"

// registerAuth wires signup/login/logout/me onto the mux. accounts is the sole authority on identity —
// storefront just proxies to it and translates the returned token into a cookie.
func (s *server) registerAuth(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		s.authProxy(w, r, "/api/auth/signup")
	})
	mux.HandleFunc("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		s.authProxy(w, r, "/api/auth/login")
	})
	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(sessionCookie); err == nil {
			_, _, _ = s.call(r.Context(), http.MethodPost, s.accountsURL+"/api/auth/logout?token="+url.QueryEscape(c.Value), nil)
		}
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		u, err := s.currentUser(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not signed in"})
			return
		}
		writeJSON(w, http.StatusOK, u)
	})
}

// authProxy calls accounts' signup/login, sets the session cookie on success, and strips the raw
// token out of the response body the browser sees (it only needs it as the cookie, not in JSON too).
func (s *server) authProxy(w http.ResponseWriter, r *http.Request, upstream string) {
	raw, status, err := s.call(r.Context(), http.MethodPost, s.accountsURL+upstream, r.Body)
	if err != nil {
		s.fail(w, r, "auth", err)
		return
	}
	if status != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(raw)
		return
	}
	var resp struct {
		UserID string `json:"userId"`
		Email  string `json:"email"`
		Name   string `json:"name"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		s.fail(w, r, "auth: decode response", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    resp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
	writeJSON(w, http.StatusOK, map[string]string{"userId": resp.UserID, "email": resp.Email, "name": resp.Name})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
}

// currentUser resolves the session cookie via accounts' verify endpoint. Called only on login-gated
// routes (checkout, order history) — not on every request, so a logged-out browse stays cheap.
func (s *server) currentUser(r *http.Request) (authUser, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return authUser{}, errNotSignedIn
	}
	raw, status, err := s.call(r.Context(), http.MethodGet, s.accountsURL+"/api/auth/verify?token="+url.QueryEscape(c.Value), nil)
	if err != nil {
		return authUser{}, err
	}
	if status != http.StatusOK {
		return authUser{}, errNotSignedIn
	}
	var u authUser
	if err := json.Unmarshal(raw, &u); err != nil {
		return authUser{}, err
	}
	return u, nil
}

type authUser struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

var errNotSignedIn = &authError{"not signed in"}

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }

// mergeAnonymousCart transfers an anonymous cart's items into the logged-in user's cart, then clears
// the anonymous one. Idempotent — merging an already-empty anonymous cart is a no-op — so it's safe to
// call on every authenticated checkout rather than tracking whether it's already happened this session.
func (s *server) mergeAnonymousCart(ctx context.Context, anonSID, userID string) {
	if anonSID == "" || anonSID == userID {
		return
	}
	raw, status, err := s.call(ctx, http.MethodGet, s.cartURL+"/api/cart/"+url.PathEscape(anonSID), nil)
	if err != nil || status != http.StatusOK {
		return
	}
	var env struct {
		Cart struct {
			Items []json.RawMessage `json:"items"`
		} `json:"cart"`
	}
	if err := json.Unmarshal(raw, &env); err != nil || len(env.Cart.Items) == 0 {
		return
	}
	for _, item := range env.Cart.Items {
		if _, _, err := s.call(ctx, http.MethodPost, s.cartURL+"/api/cart/"+url.PathEscape(userID)+"/items", jsonReader(item)); err != nil {
			telemetry.Logger.ErrorContext(ctx, "cart merge: add item failed (continuing)", "err", err)
		}
	}
	if _, _, err := s.call(ctx, http.MethodDelete, s.cartURL+"/api/cart/"+url.PathEscape(anonSID), nil); err != nil {
		telemetry.Logger.ErrorContext(ctx, "cart merge: clear anonymous cart failed (continuing)", "err", err)
	}
}

func jsonReader(v json.RawMessage) io.Reader { return bytes.NewReader(v) }
