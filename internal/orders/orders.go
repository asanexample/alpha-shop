// Package orders is the order domain + store for the bike shop. An order is a JSON document keyed by its id,
// persisted via internal/awskv (DynamoDB in-cluster, memory locally). The orders SERVICE (cmd/orders) does the
// orchestration — authorize payment, persist, emit an event — this package holds just the model + store.
package orders

import (
	"context"
	"encoding/json"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

// Status is the terminal state of a checkout.
type Status string

const (
	Placed   Status = "placed"
	Declined Status = "declined"
)

// Line is one ordered item (a price-snapshotted product).
type Line struct {
	ProductID  string `json:"productId"`
	Slug       string `json:"slug,omitempty"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Qty        int    `json:"qty"`
}

// Address is the shipping/billing address captured at checkout. Optional-in-practice today (the demo
// checkout form isn't required to fill every field), so no validation lives here — just a record of
// whatever the SPA collected.
type Address struct {
	Name    string `json:"name"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

// Order is a completed (or declined) checkout.
type Order struct {
	ID         string `json:"id"`
	SessionID  string `json:"sessionId"`
	Lines      []Line `json:"lines"`
	TotalCents int    `json:"totalCents"`
	Status     Status `json:"status"`
	PaymentID  string `json:"paymentId,omitempty"`
	Reason     string `json:"reason,omitempty"` // decline reason, when Status=Declined
	// Experience is the flagship-flagged checkout variant that ran (standard | express, ADR-099). Express
	// earns free expedited handling — a flag decision made durable on the order.
	Experience string `json:"experience,omitempty"`
	Shipping   string `json:"shipping,omitempty"`
	// ShipmentID is Bravo Dispatch's tracking number (e.g. "BD-10023") once intake has accepted this order for
	// shipping (ADR-101 cross-team integration). Empty when dispatch was unreachable — a placed order is never
	// blocked on it (best-effort, same posture as the order-placed event emit below).
	ShipmentID    string    `json:"shipmentId,omitempty"`
	Address       Address   `json:"address,omitempty"`
	PaymentMethod string    `json:"paymentMethod,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// Summary is the subset of an Order shown in a user's order-history list — enough for a list view
// without hydrating every full order.
type Summary struct {
	ID         string    `json:"id"`
	Status     Status    `json:"status"`
	TotalCents int       `json:"totalCents"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Store persists orders by id.
type Store struct{ kv awskv.Store }

// New returns an order Store over the given key-value backend.
func New(kv awskv.Store) *Store { return &Store{kv: kv} }

// Backend reports the underlying kv backend (for startup logging).
func (s *Store) Backend() string { return s.kv.Backend() }

// Save persists an order.
func (s *Store) Save(ctx context.Context, o Order) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return s.kv.Put(ctx, o.ID, b)
}

// Get returns an order by id (found=false if unknown).
func (s *Store) Get(ctx context.Context, id string) (Order, bool, error) {
	doc, found, err := s.kv.Get(ctx, id)
	if err != nil || !found {
		return Order{}, found, err
	}
	var o Order
	if err := json.Unmarshal(doc, &o); err != nil {
		return Order{}, false, err
	}
	return o, true, nil
}

// userIndexKey is a distinct key namespace in the SAME table as orders themselves (order ids are
// "ord_<hex>", so there's no collision risk) — avoids needing a second self-service resource just for
// this index.
func userIndexKey(userID string) string { return "user-orders:" + userID }

// AppendToUserIndex records an order summary under the user's history index. A read-modify-write over
// a single document — two concurrent checkouts for the same user could race and drop an index entry
// (the order itself is never lost, only its appearance in history); fine at demo scale, a real
// production version would need an actual secondary index instead.
func (s *Store) AppendToUserIndex(ctx context.Context, userID string, sum Summary) error {
	list, err := s.userIndex(ctx, userID)
	if err != nil {
		return err
	}
	list = append(list, sum)
	b, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return s.kv.Put(ctx, userIndexKey(userID), b)
}

// ListForUser returns the user's order-history index, newest first (empty if the user has no orders).
func (s *Store) ListForUser(ctx context.Context, userID string) ([]Summary, error) {
	list, err := s.userIndex(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

func (s *Store) userIndex(ctx context.Context, userID string) ([]Summary, error) {
	doc, found, err := s.kv.Get(ctx, userIndexKey(userID))
	if err != nil {
		return nil, err
	}
	if !found {
		return []Summary{}, nil
	}
	var list []Summary
	if err := json.Unmarshal(doc, &list); err != nil {
		return nil, err
	}
	return list, nil
}
