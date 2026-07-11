// Package catalogclient is the storefront BFF's typed client for the internal catalog service. Every call
// uses an otelhttp-wrapped transport (via internal/telemetry), so a storefront→catalog request propagates the
// W3C trace context and shows up as one connected distributed trace with a catalog child span.
package catalogclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/asanexample/alpha-shop/internal/catalog"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

// Client talks to the catalog service over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a client for the catalog service at baseURL (e.g. http://catalog.alpha-shop-dev.svc.cluster.local).
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: telemetry.Client()}
}

// Categories returns all catalog categories.
func (c *Client) Categories(ctx context.Context) ([]catalog.Category, error) {
	var out []catalog.Category
	return out, c.get(ctx, "/api/catalog/categories", nil, &out)
}

// Brands returns all catalog brands.
func (c *Client) Brands(ctx context.Context) ([]catalog.Brand, error) {
	var out []catalog.Brand
	return out, c.get(ctx, "/api/catalog/brands", nil, &out)
}

// Products returns the products matching the given query params (category, brand, kind, q, minPrice, …).
func (c *Client) Products(ctx context.Context, q url.Values) ([]catalog.Product, error) {
	var out struct {
		Products []catalog.Product `json:"products"`
	}
	return out.Products, c.get(ctx, "/api/catalog/products", q, &out)
}

// ProductDetail is a product plus its related items.
type ProductDetail struct {
	Product catalog.Product   `json:"product"`
	Related []catalog.Product `json:"related"`
}

// Product returns a single product (by id or slug) with related items. ok=false when not found (404).
func (c *Client) Product(ctx context.Context, idOrSlug string) (ProductDetail, bool, error) {
	var out ProductDetail
	err := c.get(ctx, "/api/catalog/products/"+url.PathEscape(idOrSlug), nil, &out)
	if err == errNotFound {
		return out, false, nil
	}
	return out, err == nil, err
}

var errNotFound = fmt.Errorf("not found")

func (c *Client) get(ctx context.Context, path string, q url.Values, dst any) error {
	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("catalog request %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog %s: status %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("catalog %s decode: %w", path, err)
	}
	return nil
}
