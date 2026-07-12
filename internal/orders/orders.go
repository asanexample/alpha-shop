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

// Order is a completed (or declined) checkout.
type Order struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"sessionId"`
	Lines      []Line    `json:"lines"`
	TotalCents int       `json:"totalCents"`
	Status     Status    `json:"status"`
	PaymentID  string    `json:"paymentId,omitempty"`
	Reason     string    `json:"reason,omitempty"` // decline reason, when Status=Declined
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
