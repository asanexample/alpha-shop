package catalog

import (
	"sort"
	"strings"
)

// Store is the in-memory catalog, seeded at construction. Read-only after New (safe for concurrent reads),
// so no locking. A real deployment would back this with a self-service resource (ADR-073); seed data keeps
// the demo deterministic.
type Store struct {
	products   []Product
	byID       map[string]Product
	bySlug     map[string]Product
	categories []Category
	brands     []Brand
	catBySlug  map[string]Category
}

// New returns a Store populated with the seeded bike-shop catalog.
func New() *Store {
	s := &Store{
		products:   seedProducts(),
		categories: seedCategories(),
		brands:     seedBrands(),
		byID:       map[string]Product{},
		bySlug:     map[string]Product{},
		catBySlug:  map[string]Category{},
	}
	for _, p := range s.products {
		s.byID[p.ID] = p
		s.bySlug[p.Slug] = p
	}
	for _, c := range s.categories {
		s.catBySlug[c.Slug] = c
	}
	return s
}

// Categories returns all categories in nav order.
func (s *Store) Categories() []Category { return s.categories }

// Brands returns all brands, alphabetically.
func (s *Store) Brands() []Brand { return s.brands }

// Category returns a category by slug (found=false if unknown).
func (s *Store) Category(slug string) (Category, bool) { c, ok := s.catBySlug[slug]; return c, ok }

// Product returns a product by its ID or slug (either works), found=false if unknown.
func (s *Store) Product(idOrSlug string) (Product, bool) {
	if p, ok := s.byID[idOrSlug]; ok {
		return p, true
	}
	p, ok := s.bySlug[idOrSlug]
	return p, ok
}

// List returns the products matching f, in a stable order (featured first, then name).
func (s *Store) List(f Filter) []Product {
	out := make([]Product, 0, len(s.products))
	for _, p := range s.products {
		if s.matches(p, f) {
			out = append(out, p)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Featured != out[j].Featured {
			return out[i].Featured // featured first
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Related returns up to n other in-stock products in the same category (for the product-detail page).
func (s *Store) Related(p Product, n int) []Product {
	out := make([]Product, 0, n)
	for _, c := range s.List(Filter{Category: p.Category}) {
		if c.ID == p.ID {
			continue
		}
		out = append(out, c)
		if len(out) == n {
			break
		}
	}
	return out
}

func (s *Store) matches(p Product, f Filter) bool {
	if f.Category != "" && p.Category != f.Category {
		return false
	}
	if f.Kind != "" {
		c, ok := s.catBySlug[p.Category]
		if !ok || c.Kind != f.Kind {
			return false
		}
	}
	if f.Brand != "" && p.Brand != f.Brand {
		return false
	}
	if f.OnSaleOnly && !p.OnSale() {
		return false
	}
	if f.FeaturedOnly && !p.Featured {
		return false
	}
	price := p.PriceCents
	if p.OnSale() {
		price = p.SalePriceCents
	}
	if f.MinPriceCents > 0 && price < f.MinPriceCents {
		return false
	}
	if f.MaxPriceCents > 0 && price > f.MaxPriceCents {
		return false
	}
	if q := strings.TrimSpace(strings.ToLower(f.Query)); q != "" {
		hay := strings.ToLower(p.Name + " " + p.Summary + " " + p.Brand + " " + p.Category)
		if !strings.Contains(hay, q) {
			return false
		}
	}
	return true
}
