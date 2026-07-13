// Command accounts is the bike-shop accounts service (team alpha, product shop, service accounts).
//
// Internal, east-west only (ClusterIP, no HTTPRoute): storefront calls it to signup/login/logout and
// to verify a session token. It's the sole authority on identity — no JWTs, no signing secret shared
// with any other service. Users + sessions persist in the self-service DynamoDB tables published as
// USERS_TABLE / SESSIONS_TABLE in the accounts-resources ConfigMap (ADR-073); absent -> in-memory
// (local dev). Instrumented via internal/telemetry.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/asanexample/alpha-shop/internal/accounts"
	"github.com/asanexample/alpha-shop/internal/awskv"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

func newMux(store *accounts.Store) *http.ServeMux {
	mux := http.NewServeMux()
	log := telemetry.Logger

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}
		if !strings.Contains(req.Email, "@") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "a valid email is required"})
			return
		}
		if len(req.Password) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
			return
		}
		u, sess, err := store.Signup(r.Context(), req.Email, req.Password, req.Name)
		if errors.Is(err, accounts.ErrEmailTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "an account with that email already exists"})
			return
		}
		if err != nil {
			fail(w, r, "signup", err)
			return
		}
		log.InfoContext(r.Context(), "account created", "userId", u.ID)
		writeAuth(w, u, sess)
	})

	mux.HandleFunc("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}
		u, sess, err := store.Login(r.Context(), req.Email, req.Password)
		if errors.Is(err, accounts.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		if err != nil {
			fail(w, r, "login", err)
			return
		}
		log.InfoContext(r.Context(), "login", "userId", u.ID)
		writeAuth(w, u, sess)
	})

	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
			return
		}
		if err := store.Logout(r.Context(), token); err != nil {
			fail(w, r, "logout", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /api/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token is required"})
			return
		}
		u, err := store.Verify(r.Context(), token)
		if errors.Is(err, accounts.ErrSessionInvalid) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired session"})
			return
		}
		if err != nil {
			fail(w, r, "verify", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"userId": u.ID, "email": u.Email, "name": u.Name})
	})

	return mux
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-accounts")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	users, err := awskv.Open(ctx, os.Getenv("USERS_TABLE"))
	if err != nil {
		telemetry.Logger.Error("users kv init failed", "err", err)
		os.Exit(1)
	}
	sessions, err := awskv.Open(ctx, os.Getenv("SESSIONS_TABLE"))
	if err != nil {
		telemetry.Logger.Error("sessions kv init failed", "err", err)
		os.Exit(1)
	}
	store := accounts.New(users, sessions)
	telemetry.Logger.Info("accounts store", "backend", store.Backend())

	srv := &http.Server{Addr: getenv("ADDR", ":8080"), Handler: telemetry.WrapHandler(newMux(store), "http.server"), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-accounts", "addr", srv.Addr)
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

// writeAuth returns the user plus the session token (the caller sets it as the cookie value).
func writeAuth(w http.ResponseWriter, u accounts.User, sess accounts.Session) {
	writeJSON(w, http.StatusOK, map[string]any{
		"userId": u.ID,
		"email":  u.Email,
		"name":   u.Name,
		"token":  sess.Token,
	})
}

func fail(w http.ResponseWriter, r *http.Request, what string, err error) {
	telemetry.Logger.ErrorContext(r.Context(), "accounts error", "what", what, "err", err)
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
