// Command orders is the bike-shop orders service (team alpha, product shop, service orders).
//
// Internal, east-west only (ClusterIP, no HTTPRoute): the storefront BFF calls it to place an order. It
// orchestrates the checkout — authorize payment (east-west call to the payment service), persist the order to
// the self-service DynamoDB table (ORDERS_TABLE), and emit an order-placed event to the self-service SQS queue
// (EVENTS_QUEUE_URL) — all in one distributed trace: storefront → orders → payment. Backends fall back to
// memory / no-op when the self-service ConfigMap keys are unset (local dev).
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
	"github.com/asanexample/alpha-shop/internal/awsqueue"
	"github.com/asanexample/alpha-shop/internal/orders"
	"github.com/asanexample/alpha-shop/internal/payment"
	"github.com/asanexample/alpha-shop/internal/paymentclient"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

type server struct {
	store  *orders.Store
	events awsqueue.Publisher
	pay    *paymentclient.Client
	now    func() time.Time
}

type placeRequest struct {
	SessionID string        `json:"sessionId"`
	Lines     []orders.Line `json:"lines"`
	Card      string        `json:"card,omitempty"`
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	log := telemetry.Logger

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/orders", func(w http.ResponseWriter, r *http.Request) {
		var req placeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}
		total := 0
		for _, l := range req.Lines {
			total += l.PriceCents * l.Qty
		}
		if len(req.Lines) == 0 || total <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cart is empty"})
			return
		}

		o := orders.Order{
			ID:         orderID(),
			SessionID:  req.SessionID,
			Lines:      req.Lines,
			TotalCents: total,
			CreatedAt:  s.now(),
		}

		// Authorize payment east-west (storefront → orders → payment in one trace).
		res, err := s.pay.Charge(r.Context(), payment.ChargeRequest{
			OrderRef:    o.ID,
			AmountCents: total,
			Currency:    "usd",
			Card:        req.Card,
		})
		if err != nil {
			log.ErrorContext(r.Context(), "payment call failed", "err", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "payment service unavailable"})
			return
		}
		o.PaymentID = res.PaymentID
		if res.Status == payment.Approved {
			o.Status = orders.Placed
		} else {
			o.Status, o.Reason = orders.Declined, res.Reason
		}

		if err := s.store.Save(r.Context(), o); err != nil {
			log.ErrorContext(r.Context(), "order save failed", "err", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not save order"})
			return
		}
		if o.Status == orders.Placed {
			// Order-placed event (best-effort; a failed emit must not fail a paid order).
			ev, _ := json.Marshal(map[string]any{"orderId": o.ID, "totalCents": o.TotalCents, "lines": len(o.Lines), "createdAt": o.CreatedAt})
			if err := s.events.Send(r.Context(), string(ev)); err != nil {
				log.ErrorContext(r.Context(), "order event emit failed (continuing)", "err", err)
			}
		}
		log.InfoContext(r.Context(), "order processed", "orderId", o.ID, "status", o.Status, "totalCents", o.TotalCents)
		writeJSON(w, http.StatusOK, o)
	})

	mux.HandleFunc("GET /api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		o, found, err := s.store.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	return mux
}

func main() {
	ctx := context.Background()
	shutdown, err := telemetry.Setup(ctx, "shop-orders")
	if err != nil {
		telemetry.Logger.Error("otel init failed; continuing without tracing", "err", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	kv, err := awskv.Open(ctx, os.Getenv("ORDERS_TABLE"))
	if err != nil {
		telemetry.Logger.Error("kv init failed", "err", err)
		os.Exit(1)
	}
	events, err := awsqueue.Open(ctx, os.Getenv("EVENTS_QUEUE_URL"))
	if err != nil {
		telemetry.Logger.Error("queue init failed", "err", err)
		os.Exit(1)
	}
	srv := &server{
		store:  orders.New(kv),
		events: events,
		pay:    paymentclient.New(getenv("PAYMENT_URL", "http://payment")),
		now:    func() time.Time { return time.Now().UTC() },
	}
	telemetry.Logger.Info("orders backends", "store", srv.store.Backend(), "events", events.Backend())

	httpSrv := &http.Server{Addr: getenv("ADDR", ":8080"), Handler: telemetry.WrapHandler(srv.routes(), "http.server"), ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	go func() {
		telemetry.Logger.Info("starting shop-orders", "addr", httpSrv.Addr)
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

func orderID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "ord_" + hex.EncodeToString(b)
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
