package payment

import "testing"

func TestAuthorize(t *testing.T) {
	// Normal charge approves.
	res, err := Authorize(ChargeRequest{OrderRef: "ord_1", AmountCents: 399900, Currency: "usd"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Approved || res.PaymentID == "" || res.AmountCents != 399900 {
		t.Fatalf("expected approved with id: %+v", res)
	}

	// Card ending 0000 declines (demo rule).
	if res, _ := Authorize(ChargeRequest{OrderRef: "ord_2", AmountCents: 5000, Card: "visa-0000"}); res.Status != Declined || res.Reason == "" {
		t.Fatalf("expected decline for 0000 card: %+v", res)
	}

	// Implausible amount declines.
	if res, _ := Authorize(ChargeRequest{OrderRef: "ord_3", AmountCents: 5_000_000}); res.Status != Declined {
		t.Fatalf("expected decline for huge amount: %+v", res)
	}

	// Non-positive amount errors.
	if _, err := Authorize(ChargeRequest{OrderRef: "ord_4", AmountCents: 0}); err == nil {
		t.Fatal("expected error for non-positive amount")
	}
}
