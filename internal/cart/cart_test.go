package cart

import (
	"context"
	"testing"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

func TestCartLifecycle(t *testing.T) {
	ctx := context.Background()
	s := New(awskv.NewMemory())
	sid := "sess-1"

	// Empty cart for a new session.
	c, err := s.Get(ctx, sid)
	if err != nil || c.Count() != 0 || len(c.Items) != 0 {
		t.Fatalf("new cart should be empty: %+v err=%v", c, err)
	}

	// Add two of one product, one of another.
	if _, err := s.Add(ctx, sid, Item{ProductID: "p1", Name: "Warbird", PriceCents: 399900, Qty: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(ctx, sid, Item{ProductID: "p1", PriceCents: 399900, Qty: 1}); err != nil { // merge
		t.Fatal(err)
	}
	c, _ = s.Add(ctx, sid, Item{ProductID: "p2", Name: "Lock", PriceCents: 8999, Qty: 1})

	if len(c.Items) != 2 {
		t.Fatalf("expected 2 distinct lines, got %d", len(c.Items))
	}
	if c.Count() != 3 {
		t.Fatalf("expected qty 3, got %d", c.Count())
	}
	if want := 399900*2 + 8999; c.SubtotalCents() != want {
		t.Fatalf("subtotal = %d, want %d", c.SubtotalCents(), want)
	}

	// Remove one line.
	c, _ = s.Remove(ctx, sid, "p1")
	if len(c.Items) != 1 || c.Items[0].ProductID != "p2" {
		t.Fatalf("remove failed: %+v", c.Items)
	}

	// Clear.
	if err := s.Clear(ctx, sid); err != nil {
		t.Fatal(err)
	}
	if c, _ := s.Get(ctx, sid); c.Count() != 0 {
		t.Fatalf("cart not cleared: %+v", c)
	}
}
