// Package paymentclient is the orders service's client for the internal payment service. It uses an
// otelhttp-wrapped transport (via internal/telemetry) so an orders→payment charge propagates the trace and
// shows up as a connected span in the checkout trace.
package paymentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asanexample/alpha-shop/internal/payment"
	"github.com/asanexample/alpha-shop/internal/telemetry"
)

// Client calls the payment service over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a client for the payment service at baseURL (e.g. http://payment).
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: telemetry.Client()}
}

// Charge authorizes an amount and returns the result. A business decline is a normal ChargeResult
// (Status=Declined), not an error; err is only non-nil on transport/protocol failure.
func (c *Client) Charge(ctx context.Context, req payment.ChargeRequest) (payment.ChargeResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return payment.ChargeResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/payment/charge", bytes.NewReader(body))
	if err != nil {
		return payment.ChargeResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return payment.ChargeResult{}, fmt.Errorf("payment request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return payment.ChargeResult{}, fmt.Errorf("payment: status %d", resp.StatusCode)
	}
	var out payment.ChargeResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return payment.ChargeResult{}, fmt.Errorf("payment decode: %w", err)
	}
	return out, nil
}
