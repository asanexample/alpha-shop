// Package payment is a mock payment processor for the bike shop (the orders service calls it east-west to
// "charge" a checkout). It holds no state and talks to no external gateway — it just models an authorization:
// most charges approve, a couple of believable rules decline, and each carries a small simulated latency so
// the checkout trace has a realistic payment span. Deterministic enough to demo, never real money.
package payment

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Status is the outcome of a charge authorization.
type Status string

const (
	Approved Status = "approved"
	Declined Status = "declined"
)

// ChargeRequest asks to authorize an amount for an order.
type ChargeRequest struct {
	OrderRef    string `json:"orderRef"`
	AmountCents int    `json:"amountCents"`
	Currency    string `json:"currency"`
	// Card is a demo-only descriptor (e.g. "visa-4242"); a value ending in "0000" simulates a declined card.
	Card string `json:"card,omitempty"`
	// Method is the payment method chosen at checkout (e.g. "card", "paypal", "apple_pay"). Cosmetic — the
	// same decline rules (bad card, implausible amount) apply uniformly regardless of method; this isn't a
	// distinct mock processor per method, just a recorded choice.
	Method string `json:"method,omitempty"`
}

// ChargeResult is the authorization outcome.
type ChargeResult struct {
	PaymentID   string `json:"paymentId"`
	Status      Status `json:"status"`
	AmountCents int    `json:"amountCents"`
	Method      string `json:"method,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// Authorize applies the mock authorization rules. It never errors on business decline — a decline is a normal
// ChargeResult with Status=Declined so the caller can surface it; it only errors on a malformed request.
func Authorize(req ChargeRequest) (ChargeResult, error) {
	if req.AmountCents <= 0 {
		return ChargeResult{}, fmt.Errorf("amountCents must be positive, got %d", req.AmountCents)
	}
	id, err := paymentID()
	if err != nil {
		return ChargeResult{}, err
	}
	res := ChargeResult{PaymentID: id, AmountCents: req.AmountCents, Method: req.Method, Status: Approved}

	switch {
	case len(req.Card) >= 4 && req.Card[len(req.Card)-4:] == "0000":
		res.Status, res.Reason = Declined, "card declined by issuer"
	case req.AmountCents > 2_000_000: // > $20,000 — flag an implausible cart as a fraud check
		res.Status, res.Reason = Declined, "amount exceeds authorization limit"
	}
	return res, nil
}

func paymentID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("payment id: %w", err)
	}
	return "pay_" + hex.EncodeToString(b), nil
}
