// Command payment is the bike-shop payment service (team alpha, product shop, service payment).
//
// Internal, east-west only (ClusterIP, no HTTPRoute): the orders service calls it to authorize a checkout.
// It is a mock authorizer (internal/payment) — no real gateway, no stored state — but it is fully
// instrumented, so a checkout shows up as one trace: storefront → orders → payment, with the authorization
// modelled as its own span carrying payment.status / payment.amount_cents attributes.
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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/asanexample/alpha-shop/internal/payment"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

var tracer = otel.Tracer("shop-payment")

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	log := telemetry.Logger

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/payment/charge", func(w http.ResponseWriter, r *http.Request) {
		var req payment.ChargeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}

		// Model the authorization as its own span so the checkout trace shows a distinct payment step.
		ctx, span := tracer.Start(r.Context(), "authorize")
		span.SetAttributes(
			attribute.Int("payment.amount_cents", req.AmountCents),
			attribute.String("payment.order_ref", req.OrderRef),
		)
		simulateProcessing(ctx)
		res, err := payment.Authorize(req)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.End()
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		span.SetAttributes(
			attribute.String("payment.id", res.PaymentID),
			attribute.String("payment.status", string(res.Status)),
		)
		span.End()

		log.InfoContext(r.Context(), "charge authorized", "status", res.Status, "amountCents", res.AmountCents, "paymentId", res.PaymentID)
		writeJSON(w, http.StatusOK, res)
	})

	return mux
}

// simulateProcessing adds a small, believable authorization latency (a real gateway round-trip) so the
// payment span has visible duration in the trace. Cancellable via the request context.
func simulateProcessing(ctx context.Context) {
	t := time.NewTimer(45 * time.Millisecond)
	defer t.Stop()
	select {
	case <-t.C:
	case <-ctx.Done():
	}
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-payment")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	srv := &http.Server{Addr: getenv("ADDR", ":8080"), Handler: telemetry.WrapHandler(newMux(), "http.server"), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-payment", "addr", srv.Addr)
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
