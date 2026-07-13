// Command orders is the bike-shop orders service (team alpha, product shop, service orders).
//
// Internal, east-west only (ClusterIP, no HTTPRoute): the storefront BFF calls it to place an order. It
// orchestrates the checkout — authorize payment (east-west call to the payment service), ask Bravo Dispatch's
// intake to kick off a real shipment (cross-team east-west, ADR-101), persist the order to the self-service
// DynamoDB table (ORDERS_TABLE), and emit an order-placed event to the self-service SQS queue
// (EVENTS_QUEUE_URL) — all in one distributed trace: storefront → orders → payment (+ orders → bravo
// intake → shipments/dispatch-worker, a trace that now spans two teams). Backends fall back to memory / no-op
// when the self-service ConfigMap keys are unset (local dev); the dispatch call is skipped entirely when
// DISPATCH_URL is unset (any stage without the ServiceGrant-backed dependency declared).
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
	"github.com/asanexample/alpha-shop/internal/dispatchclient"
	"github.com/asanexample/alpha-shop/internal/orders"
	"github.com/asanexample/alpha-shop/internal/payment"
	"github.com/asanexample/alpha-shop/internal/paymentclient"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

type server struct {
	store    *orders.Store
	events   awsqueue.Publisher
	pay      *paymentclient.Client
	dispatch *dispatchclient.Client
	now      func() time.Time
}

type placeRequest struct {
	SessionID     string         `json:"sessionId"`
	Lines         []orders.Line  `json:"lines"`
	Card          string         `json:"card,omitempty"`
	Experience    string         `json:"experience,omitempty"` // flagship checkout variant (standard | express)
	Address       orders.Address `json:"address,omitempty"`
	PaymentMethod string         `json:"paymentMethod,omitempty"`
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
			ID:            orderID(),
			SessionID:     req.SessionID,
			Lines:         req.Lines,
			TotalCents:    total,
			Experience:    req.Experience,
			Shipping:      "Standard (5–7 days)",
			Address:       req.Address,
			PaymentMethod: req.PaymentMethod,
			CreatedAt:     s.now(),
		}
		if req.Experience == "express" {
			o.Shipping = "Express — free expedited (1–2 days)"
		}

		// Authorize payment east-west (storefront → orders → payment in one trace).
		res, err := s.pay.Charge(r.Context(), payment.ChargeRequest{
			OrderRef:    o.ID,
			AmountCents: total,
			Currency:    "usd",
			Card:        req.Card,
			Method:      req.PaymentMethod,
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

		if o.Status == orders.Placed && s.dispatch.Enabled() {
			// Cross-team east-west call (ADR-101): ask Bravo Dispatch's intake to create + route a real
			// shipment. Best-effort — a failed or unreachable dispatch must not fail an already-paid order;
			// it just means no ShipmentID this time (the demo equivalent of "we'll email you a tracking
			// number shortly"). Use the real address when the checkout form captured one; synthesized demo
			// values otherwise (same spirit as the Card field above), since address capture isn't required.
			recipient, destination := "Alpha Bikes customer "+o.SessionID, "Customer Address on File"
			if o.Address.Name != "" {
				recipient = o.Address.Name
			}
			if o.Address.Line1 != "" {
				destination = o.Address.Line1 + ", " + o.Address.City + ", " + o.Address.State + " " + o.Address.Zip
			}
			sh, err := s.dispatch.CreateShipment(r.Context(), dispatchclient.CreateShipmentRequest{
				Recipient:   recipient,
				Origin:      "Alpha Bikes Warehouse",
				Destination: destination,
			})
			if err != nil {
				log.ErrorContext(r.Context(), "dispatch shipment request failed (continuing)", "orderId", o.ID, "err", err)
			} else {
				o.ShipmentID = sh.ID
			}
		}

		if err := s.store.Save(r.Context(), o); err != nil {
			log.ErrorContext(r.Context(), "order save failed", "err", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not save order"})
			return
		}
		// Record it in the user's order-history index (both placed and declined attempts — a customer
		// should be able to see a failed payment in their history, not just successful orders). Best-effort:
		// a failed index write must not fail an already-saved order, same posture as the event emit below.
		if err := s.store.AppendToUserIndex(r.Context(), o.SessionID, orders.Summary{ID: o.ID, Status: o.Status, TotalCents: o.TotalCents, CreatedAt: o.CreatedAt}); err != nil {
			log.ErrorContext(r.Context(), "order history index update failed (continuing)", "err", err)
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

	// GET /api/orders?userId=... — order history for the signed-in user (storefront resolves + supplies
	// userId after verifying the session; orders itself still knows nothing about sessions/cookies).
	mux.HandleFunc("GET /api/orders", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "userId is required"})
			return
		}
		list, err := s.store.ListForUser(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"orders": list})
	})

	// GET /api/orders/{id}?userId=... — single-order lookup, now ownership-checked: an order carries a
	// real name+address (PII) since this service gained address capture, so a 404 on mismatch (not just
	// "any known id works") matters here in a way it didn't when an order was just cart contents + a price.
	mux.HandleFunc("GET /api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "userId is required"})
			return
		}
		o, found, err := s.store.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if !found || o.SessionID != userID {
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
		store:    orders.New(kv),
		events:   events,
		pay:      paymentclient.New(getenv("PAYMENT_URL", "http://payment")),
		dispatch: dispatchclient.New(os.Getenv("DISPATCH_URL")), // empty = no ServiceGrant-backed dependency in this stage
		now:      func() time.Time { return time.Now().UTC() },
	}
	telemetry.Logger.Info("orders backends", "store", srv.store.Backend(), "events", events.Backend(), "dispatch", srv.dispatch.Enabled())

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
