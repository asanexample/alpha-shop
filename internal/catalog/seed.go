package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// This is the seeded bike-shop catalog — a small, believable dataset (à la a local independent shop). Product
// images are rendered by the storefront as deterministic placeholders keyed on the product ID, so there are no
// binary assets to ship. Prices are in USD cents. Categories/brands are small, stable lists kept as Go
// literals; the product list itself is big enough (100+) that it lives in data/products.json instead, loaded
// via go:embed so the binary stays a single static artifact.

//go:embed data/products.json
var productsJSON []byte

func seedProducts() []Product {
	var products []Product
	if err := json.Unmarshal(productsJSON, &products); err != nil {
		panic(fmt.Sprintf("catalog: embedded products.json is invalid: %v", err))
	}
	return products
}

func seedCategories() []Category {
	return []Category{
		{Slug: "road", Name: "Road Bikes", Kind: KindBikes, Blurb: "Fast on the tarmac — endurance to race."},
		{Slug: "gravel", Name: "Gravel Bikes", Kind: KindBikes, Blurb: "One bike, every surface."},
		{Slug: "mountain", Name: "Mountain Bikes", Kind: KindBikes, Blurb: "Trail, all-mountain, and hardtail."},
		{Slug: "commuter", Name: "Commuter & Urban", Kind: KindBikes, Blurb: "Get to work, rain or shine."},
		{Slug: "kids", Name: "Kids' Bikes", Kind: KindBikes, Blurb: "First pedals to first trails."},
		{Slug: "ebike-city", Name: "Electric City", Kind: KindEbikes, Blurb: "Pedal-assist for the daily ride."},
		{Slug: "ebike-mountain", Name: "Electric Mountain", Kind: KindEbikes, Blurb: "More trail, less shuttle."},
		{Slug: "drivetrain", Name: "Drivetrain", Kind: KindComponents, Blurb: "Shift and stop with confidence."},
		{Slug: "wheelsets", Name: "Wheelsets", Kind: KindWheels, Blurb: "Lighter, stiffer, faster."},
		{Slug: "tires", Name: "Tires & Tubes", Kind: KindTires, Blurb: "Grip for every condition."},
		{Slug: "accessories", Name: "Accessories", Kind: KindAccessories, Blurb: "Lights, locks, bags, and tools."},
		{Slug: "apparel", Name: "Apparel & Helmets", Kind: KindApparel, Blurb: "Ride comfortable, ride safe."},
	}
}

func seedBrands() []Brand {
	return []Brand{
		{Slug: "all-city", Name: "All-City"},
		{Slug: "cannondale", Name: "Cannondale"},
		{Slug: "giant", Name: "Giant"},
		{Slug: "kona", Name: "Kona"},
		{Slug: "marin", Name: "Marin"},
		{Slug: "salsa", Name: "Salsa"},
		{Slug: "specialized", Name: "Specialized"},
		{Slug: "surly", Name: "Surly"},
		{Slug: "trek", Name: "Trek"},
		{Slug: "yeti", Name: "Yeti"},
	}
}
