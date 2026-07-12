// Package dispatchclient is the orders service's client for team bravo's Bravo Dispatch product — the
// cross-team hop that kicks off a real shipment when a bike-shop order is placed. It uses an otelhttp-wrapped
// transport (via internal/telemetry) so orders→intake propagates the trace and shows up as a connected span
// spanning both teams' services, the same as the in-team payment/cart calls. The network path (both CNP
// halves, in the alpha-shop-dev AND bravo-dispatch-dev namespaces) is entirely governed by a ServiceGrant
// (ADR-101) — bravo consents in gitops/grants/bravo/, alpha declares intent on its own XEnvironment claim's
// spec.dependencies; this client hand-authors none of that.
package dispatchclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asanexample/alpha-shop/internal/telemetry"
)

// Client calls Bravo Dispatch's intake service over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a client for intake at baseURL (e.g. http://intake.bravo-dispatch-dev). Empty baseURL is a
// valid, deliberate "no dispatch integration in this stage" state — see Enabled().
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: telemetry.Client()}
}

// Enabled reports whether a dispatch target is configured at all.
func (c *Client) Enabled() bool { return c.baseURL != "" }

// CreateShipmentRequest mirrors intake's POST /shipments body (cmd/intake/main.go, bravo-dispatch repo).
type CreateShipmentRequest struct {
	Recipient   string `json:"recipient"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

// Shipment mirrors the subset of intake's response this caller needs — just enough to surface a tracking
// number, not bravo's full shipment record shape.
type Shipment struct {
	ID     string `json:"id"` // the tracking number, e.g. "BD-10023"
	Status string `json:"status"`
}

// CreateShipment asks intake to create + route a shipment for a placed order. err is non-nil only on
// transport/protocol failure (unreachable, non-2xx, bad body) — callers treat any error as "no shipment yet",
// never as a reason to fail the order itself (see cmd/orders/main.go).
func (c *Client) CreateShipment(ctx context.Context, req CreateShipmentRequest) (Shipment, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Shipment{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/shipments", bytes.NewReader(body))
	if err != nil {
		return Shipment{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return Shipment{}, fmt.Errorf("dispatch intake request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		return Shipment{}, fmt.Errorf("dispatch intake: status %d", resp.StatusCode)
	}
	var out Shipment
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Shipment{}, fmt.Errorf("dispatch intake decode: %w", err)
	}
	return out, nil
}

// ProbeResult is the raw outcome of an arbitrary method/path call to intake — used only by Probe, to observe
// the ServiceGrant's L7 CiliumNetworkPolicy enforcing (or not) at the network layer, distinct from a normal
// application-level response.
type ProbeResult struct {
	StatusCode int    `json:"statusCode,omitempty"`
	Body       string `json:"body,omitempty"`
	Err        string `json:"err,omitempty"`
}

// Probe sends an arbitrary method+path to intake and reports the raw outcome — including a transport-level
// failure (connection reset/timeout), which is what a CiliumNetworkPolicy L7 rejection at the network layer
// looks like from the caller's side, as opposed to a normal HTTP error response from intake's own app code.
// ADR-101 verification tool ONLY (cmd/orders' /debug/dispatch-probe route) — ProbeResult.Err is populated on
// any transport failure; a non-empty StatusCode+Body means the request reached intake's app.
func (c *Client) Probe(ctx context.Context, method, path string) ProbeResult {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return ProbeResult{Err: err.Error()}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return ProbeResult{Err: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return ProbeResult{StatusCode: resp.StatusCode, Body: string(body)}
}
