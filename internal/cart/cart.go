// Package cart is the shopping-cart domain + store for the bike shop. A cart is a small JSON document keyed
// by the caller's session id, persisted via internal/awskv (DynamoDB in-cluster, memory locally). Line items
// carry a price snapshot so the cart total is stable even if the catalog price changes later.
package cart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

// Item is one line in the cart. PriceCents is snapshotted at add-time.
type Item struct {
	ProductID  string `json:"productId"`
	Slug       string `json:"slug,omitempty"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Qty        int    `json:"qty"`
}

// Cart is a session's set of line items.
type Cart struct {
	SessionID  string    `json:"sessionId"`
	Items      []Item    `json:"items"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// SubtotalCents is the sum of price×qty across items.
func (c Cart) SubtotalCents() int {
	t := 0
	for _, it := range c.Items {
		t += it.PriceCents * it.Qty
	}
	return t
}

// Count is the total item quantity.
func (c Cart) Count() int {
	n := 0
	for _, it := range c.Items {
		n += it.Qty
	}
	return n
}

// Store persists carts by session id.
type Store struct {
	kv  awskv.Store
	now func() time.Time
}

// New returns a cart Store over the given key-value backend.
func New(kv awskv.Store) *Store { return &Store{kv: kv, now: func() time.Time { return time.Now().UTC() }} }

// Backend reports the underlying kv backend (for startup logging).
func (s *Store) Backend() string { return s.kv.Backend() }

// Get returns the session's cart (an empty cart if none exists yet).
func (s *Store) Get(ctx context.Context, sessionID string) (Cart, error) {
	doc, found, err := s.kv.Get(ctx, sessionID)
	if err != nil {
		return Cart{}, err
	}
	if !found {
		return Cart{SessionID: sessionID, Items: []Item{}}, nil
	}
	var c Cart
	if err := json.Unmarshal(doc, &c); err != nil {
		return Cart{}, err
	}
	if c.Items == nil {
		c.Items = []Item{}
	}
	c.SessionID = sessionID
	return c, nil
}

// Add adds qty of an item (merging into an existing line for the same product). qty<=0 is treated as 1.
func (s *Store) Add(ctx context.Context, sessionID string, item Item) (Cart, error) {
	if item.Qty <= 0 {
		item.Qty = 1
	}
	c, err := s.Get(ctx, sessionID)
	if err != nil {
		return Cart{}, err
	}
	found := false
	for i := range c.Items {
		if c.Items[i].ProductID == item.ProductID {
			c.Items[i].Qty += item.Qty
			found = true
			break
		}
	}
	if !found {
		c.Items = append(c.Items, item)
	}
	return c, s.save(ctx, &c)
}

// Remove deletes a product's line entirely.
func (s *Store) Remove(ctx context.Context, sessionID, productID string) (Cart, error) {
	c, err := s.Get(ctx, sessionID)
	if err != nil {
		return Cart{}, err
	}
	kept := c.Items[:0]
	for _, it := range c.Items {
		if it.ProductID != productID {
			kept = append(kept, it)
		}
	}
	c.Items = kept
	return c, s.save(ctx, &c)
}

// Clear empties the session's cart.
func (s *Store) Clear(ctx context.Context, sessionID string) error {
	return s.kv.Delete(ctx, sessionID)
}

func (s *Store) save(ctx context.Context, c *Cart) error {
	c.UpdatedAt = s.now()
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return s.kv.Put(ctx, c.SessionID, b)
}
