package orders

import (
	"context"
	"testing"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

func TestSaveGet(t *testing.T) {
	ctx := context.Background()
	s := New(awskv.NewMemory())

	o := Order{ID: "ord_1", SessionID: "user@example.com", TotalCents: 5000, Status: Placed, CreatedAt: time.Now().UTC()}
	if err := s.Save(ctx, o); err != nil {
		t.Fatal(err)
	}

	got, found, err := s.Get(ctx, "ord_1")
	if err != nil || !found || got.ID != o.ID {
		t.Fatalf("get = %+v, found=%v, err=%v", got, found, err)
	}

	if _, found, err := s.Get(ctx, "not-a-real-order"); err != nil || found {
		t.Fatalf("expected not-found for unknown id, got found=%v err=%v", found, err)
	}
}

func TestUserOrderIndex(t *testing.T) {
	ctx := context.Background()
	s := New(awskv.NewMemory())
	userID := "rider@example.com"

	// No orders yet.
	list, err := s.ListForUser(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty history, got %+v", list)
	}

	first := Summary{ID: "ord_1", Status: Placed, TotalCents: 1000, CreatedAt: time.Now().UTC()}
	second := Summary{ID: "ord_2", Status: Declined, TotalCents: 2000, CreatedAt: time.Now().UTC().Add(time.Minute)}
	if err := s.AppendToUserIndex(ctx, userID, first); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendToUserIndex(ctx, userID, second); err != nil {
		t.Fatal(err)
	}

	list, err = s.ListForUser(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(list))
	}
	// Newest first.
	if list[0].ID != "ord_2" || list[1].ID != "ord_1" {
		t.Fatalf("expected newest-first order, got %+v", list)
	}

	// A different user's history is separate.
	other, err := s.ListForUser(ctx, "someone-else@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Fatalf("expected empty history for a different user, got %+v", other)
	}
}
