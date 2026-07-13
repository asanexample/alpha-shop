package catalog

import (
	"context"
	"testing"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(context.Background(), awskv.NewMemory())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestStoreSeeded(t *testing.T) {
	s := newTestStore(t)
	if len(s.Categories()) == 0 || len(s.Brands()) == 0 {
		t.Fatal("expected seeded categories and brands")
	}
	if got := s.List(Filter{}); len(got) == 0 {
		t.Fatal("expected seeded products")
	}
	// Every product must reference a known category and brand — a seed-integrity guard.
	brands := map[string]bool{}
	for _, b := range s.Brands() {
		brands[b.Slug] = true
	}
	for _, p := range s.List(Filter{}) {
		if _, ok := s.Category(p.Category); !ok {
			t.Errorf("product %s references unknown category %q", p.ID, p.Category)
		}
		if !brands[p.Brand] {
			t.Errorf("product %s references unknown brand %q", p.ID, p.Brand)
		}
	}
}

func TestFilter(t *testing.T) {
	s := newTestStore(t)

	gravel := s.List(Filter{Category: "gravel"})
	if len(gravel) == 0 {
		t.Fatal("expected gravel products")
	}
	for _, p := range gravel {
		if p.Category != "gravel" {
			t.Errorf("category filter leaked %s (%s)", p.ID, p.Category)
		}
	}

	// Kind filter spans the categories in a section.
	bikes := s.List(Filter{Kind: KindBikes})
	ebikes := s.List(Filter{Kind: KindEbikes})
	if len(bikes) == 0 || len(ebikes) == 0 || len(bikes) == len(s.List(Filter{})) {
		t.Fatalf("kind filter looks wrong: bikes=%d ebikes=%d total=%d", len(bikes), len(ebikes), len(s.List(Filter{})))
	}

	// On-sale filter returns only discounted items.
	for _, p := range s.List(Filter{OnSaleOnly: true}) {
		if !p.OnSale() {
			t.Errorf("onSale filter returned non-sale product %s", p.ID)
		}
	}

	// Free-text query matches brand.
	if got := s.List(Filter{Query: "surly"}); len(got) == 0 {
		t.Error("expected query 'surly' to match products")
	}

	// Price ceiling (dollars→cents already applied by caller; here cents).
	for _, p := range s.List(Filter{MaxPriceCents: 100000}) {
		price := p.PriceCents
		if p.OnSale() {
			price = p.SalePriceCents
		}
		if price > 100000 {
			t.Errorf("maxPrice filter leaked %s at %d", p.ID, price)
		}
	}
}

func TestListPaged(t *testing.T) {
	s := newTestStore(t)
	all := s.List(Filter{})

	// Defaults: page 1, defaultPerPage-sized (or the full set if smaller).
	first := s.ListPaged(Filter{})
	if first.Total != len(all) || first.Page != 1 || first.PerPage != defaultPerPage {
		t.Fatalf("unexpected defaults: %+v (want total=%d)", first, len(all))
	}
	wantFirstLen := defaultPerPage
	if wantFirstLen > len(all) {
		wantFirstLen = len(all)
	}
	if len(first.Products) != wantFirstLen {
		t.Fatalf("expected %d products on page 1, got %d", wantFirstLen, len(first.Products))
	}

	// Paging through never repeats a product and covers the whole filtered set.
	seen := map[string]bool{}
	for page := 1; ; page++ {
		listing := s.ListPaged(Filter{Page: page, PerPage: 10})
		if len(listing.Products) == 0 {
			break
		}
		for _, p := range listing.Products {
			if seen[p.ID] {
				t.Fatalf("product %s seen on more than one page", p.ID)
			}
			seen[p.ID] = true
		}
		if page > len(all) { // safety valve against an infinite loop on a bug
			t.Fatal("paging did not terminate")
		}
	}
	if len(seen) != len(all) {
		t.Fatalf("paged through %d products, want %d", len(seen), len(all))
	}

	// A page past the end is empty, not an error, and still reports the true total.
	past := s.ListPaged(Filter{Page: 9999, PerPage: 10})
	if len(past.Products) != 0 || past.Total != len(all) {
		t.Fatalf("expected empty page past the end, got %+v", past)
	}

	// PerPage is capped.
	capped := s.ListPaged(Filter{PerPage: 10_000})
	if capped.PerPage != maxPerPage {
		t.Fatalf("expected PerPage capped at %d, got %d", maxPerPage, capped.PerPage)
	}
}

func TestNewSeedsOnceThenPersists(t *testing.T) {
	ctx := context.Background()
	kv := awskv.NewMemory()

	first, err := New(ctx, kv)
	if err != nil {
		t.Fatalf("New (seed): %v", err)
	}
	seeded := first.List(Filter{})
	if len(seeded) == 0 {
		t.Fatal("expected the first New to seed products into kv")
	}

	// Simulate a second boot (e.g. a second replica, or a restart) against the same table: it must load
	// the persisted document rather than re-seeding, and see the identical catalog.
	if _, found, err := kv.Get(ctx, productsKey); err != nil || !found {
		t.Fatalf("expected products document persisted in kv, found=%v err=%v", found, err)
	}

	second, err := New(ctx, kv)
	if err != nil {
		t.Fatalf("New (reload): %v", err)
	}
	if got := second.List(Filter{}); len(got) != len(seeded) {
		t.Fatalf("second New saw %d products, want %d (same as first boot)", len(got), len(seeded))
	}

	// A third New against a distinct empty store must NOT see the first store's seeded data (no shared
	// global state) and must independently seed its own copy.
	third, err := New(ctx, awskv.NewMemory())
	if err != nil {
		t.Fatalf("New (independent seed): %v", err)
	}
	if got := third.List(Filter{}); len(got) != len(seeded) {
		t.Fatalf("independently-seeded store saw %d products, want %d", len(got), len(seeded))
	}
}

func TestProductLookupAndRelated(t *testing.T) {
	s := newTestStore(t)
	p, ok := s.Product("salsa-warbird-c-grx")
	if !ok {
		t.Fatal("expected lookup by slug")
	}
	if byID, ok := s.Product(p.ID); !ok || byID.Slug != p.Slug {
		t.Fatal("expected lookup by ID to match")
	}
	for _, r := range s.Related(p, 4) {
		if r.ID == p.ID {
			t.Error("related must not include the product itself")
		}
		if r.Category != p.Category {
			t.Errorf("related product %s is a different category", r.ID)
		}
	}
}
