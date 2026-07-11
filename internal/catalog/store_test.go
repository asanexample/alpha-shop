package catalog

import "testing"

func TestStoreSeeded(t *testing.T) {
	s := New()
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
	s := New()

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

func TestProductLookupAndRelated(t *testing.T) {
	s := New()
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
