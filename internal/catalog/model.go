// Package catalog is the bike-shop product catalog: the domain model, an in-memory seeded store, and the
// query/filter logic behind the catalog service's read API. It has no platform dependencies — a plain,
// deterministic dataset so the shop is believable and the demo is reproducible (no external DB for a tenant;
// ADR-073 self-service resources would back a real store, but seed data keeps the browse path simple).
package catalog

// Kind groups categories into the shop's top-level nav sections.
type Kind string

const (
	KindBikes       Kind = "bikes"
	KindEbikes      Kind = "ebikes"
	KindComponents  Kind = "components"
	KindWheels      Kind = "wheels"
	KindTires       Kind = "tires"
	KindAccessories Kind = "accessories"
	KindApparel     Kind = "apparel"
)

// Category is a browsable section of the shop (e.g. "Gravel Bikes").
type Category struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Kind Kind   `json:"kind"`
	// Blurb is a short line shown on the category landing.
	Blurb string `json:"blurb,omitempty"`
}

// Brand is a manufacturer the shop carries.
type Brand struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Product is a single catalog item. Prices are in whole USD cents to avoid float drift. SalePriceCents is 0
// when the item is not on sale.
type Product struct {
	ID             string            `json:"id"`
	Slug           string            `json:"slug"`
	Name           string            `json:"name"`
	Brand          string            `json:"brand"`    // Brand.Slug
	Category       string            `json:"category"` // Category.Slug
	PriceCents     int               `json:"priceCents"`
	SalePriceCents int               `json:"salePriceCents,omitempty"`
	Summary        string            `json:"summary"`
	Description     string           `json:"description"`
	Image          string            `json:"image"` // relative asset path served by the storefront
	Sizes          []string          `json:"sizes,omitempty"`
	Specs          map[string]string `json:"specs,omitempty"`
	InStock        bool              `json:"inStock"`
	Featured       bool              `json:"featured,omitempty"`
}

// OnSale reports whether the product carries a sale price below its list price.
func (p Product) OnSale() bool { return p.SalePriceCents > 0 && p.SalePriceCents < p.PriceCents }

// Filter is the set of query constraints the product listing supports (all optional; zero value = no filter).
type Filter struct {
	Category      string // Category.Slug
	Kind          Kind   // top-level section
	Brand         string // Brand.Slug
	Query         string // free-text over name/summary/brand
	MinPriceCents int
	MaxPriceCents int
	OnSaleOnly    bool
	FeaturedOnly  bool
}
