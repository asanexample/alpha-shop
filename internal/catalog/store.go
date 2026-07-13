package catalog

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

// productsKey is the single document holding the whole product list, in the same self-service
// DynamoDB table (ADR-073) every other stateful shop service already uses via internal/awskv.
const productsKey = "products"

// Store is the catalog, loaded via kv at construction and cached in memory (read-only after New, safe
// for concurrent reads — no locking). Categories/brands are a small, stable nav taxonomy kept as Go
// code (config, not data); products are the shop's actual catalog data and live in kv.
type Store struct {
	kv         awskv.Store
	products   []Product
	byID       map[string]Product
	bySlug     map[string]Product
	categories []Category
	brands     []Brand
	catBySlug  map[string]Category
}

// New returns a Store backed by kv. On an empty table (first boot in-cluster, or every boot against the
// in-memory local-dev backend) it seeds the products document from the embedded JSON and persists it;
// otherwise it loads the products already there — so a real deployment reads its catalog from Dynamo like
// every other stateful service, instead of carrying it compiled into the binary.
func New(ctx context.Context, kv awskv.Store) (*Store, error) {
	products, err := loadOrSeedProducts(ctx, kv)
	if err != nil {
		return nil, err
	}
	s := &Store{
		kv:         kv,
		products:   products,
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
	return s, nil
}

func loadOrSeedProducts(ctx context.Context, kv awskv.Store) ([]Product, error) {
	if doc, found, err := kv.Get(ctx, productsKey); err != nil {
		return nil, err
	} else if found {
		var products []Product
		if err := json.Unmarshal(doc, &products); err != nil {
			return nil, err
		}
		return products, nil
	}
	products := seedProducts()
	doc, err := json.Marshal(products)
	if err != nil {
		return nil, err
	}
	if err := kv.Put(ctx, productsKey, doc); err != nil {
		return nil, err
	}
	return products, nil
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

// defaultPerPage/maxPerPage bound ListPaged when the caller omits or abuses PerPage.
const (
	defaultPerPage = 24
	maxPerPage     = 100
)

// ListPaged is List plus pagination — the entry point for the browse/search HTTP API. Page is 1-indexed;
// Page/PerPage <= 0 fall back to page 1 / defaultPerPage, and PerPage is capped at maxPerPage.
func (s *Store) ListPaged(f Filter) Listing {
	all := s.List(f)
	page, perPage := f.Page, f.PerPage
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	start := (page - 1) * perPage
	if start >= len(all) {
		return Listing{Products: []Product{}, Total: len(all), Page: page, PerPage: perPage}
	}
	end := start + perPage
	if end > len(all) {
		end = len(all)
	}
	return Listing{Products: all[start:end], Total: len(all), Page: page, PerPage: perPage}
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
