package accounts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
)

func newTestStore() *Store {
	return New(awskv.NewMemory(), awskv.NewMemory())
}

func TestSignupLoginVerifyLogout(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()

	u, sess, err := s.Signup(ctx, "Rider@Example.com ", "hunter22", "Rider")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "rider@example.com" || sess.Token == "" {
		t.Fatalf("unexpected signup result: %+v %+v", u, sess)
	}

	// Duplicate signup (any case/whitespace variant) is rejected.
	if _, _, err := s.Signup(ctx, "  RIDER@example.com", "somethingelse", "Rider2"); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}

	// Wrong password rejected.
	if _, _, err := s.Login(ctx, "rider@example.com", "wrongpass"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	// Correct login mints a fresh session.
	_, loginSess, err := s.Login(ctx, "rider@example.com", "hunter22")
	if err != nil {
		t.Fatal(err)
	}
	if loginSess.Token == sess.Token {
		t.Fatal("expected a distinct session token from login")
	}

	// Both the signup and login sessions verify.
	if got, err := s.Verify(ctx, sess.Token); err != nil || got.ID != u.ID {
		t.Fatalf("verify(signup token) = %+v, %v", got, err)
	}
	if got, err := s.Verify(ctx, loginSess.Token); err != nil || got.ID != u.ID {
		t.Fatalf("verify(login token) = %+v, %v", got, err)
	}

	// Logout invalidates only that token.
	if err := s.Logout(ctx, sess.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Verify(ctx, sess.Token); !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected ErrSessionInvalid after logout, got %v", err)
	}
	if _, err := s.Verify(ctx, loginSess.Token); err != nil {
		t.Fatalf("other session should remain valid: %v", err)
	}

	// Unknown token is invalid.
	if _, err := s.Verify(ctx, "not-a-real-token"); !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected ErrSessionInvalid for unknown token, got %v", err)
	}
}

func TestExpiredSession(t *testing.T) {
	ctx := context.Background()
	s := newTestStore()
	frozen := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return frozen }

	_, sess, err := s.Signup(ctx, "old@example.com", "hunter22", "Old")
	if err != nil {
		t.Fatal(err)
	}

	// Still valid just before expiry.
	s.now = func() time.Time { return frozen.Add(sessionTTL - time.Second) }
	if _, err := s.Verify(ctx, sess.Token); err != nil {
		t.Fatalf("expected still-valid session, got %v", err)
	}

	// Expired just after.
	s.now = func() time.Time { return frozen.Add(sessionTTL + time.Second) }
	if _, err := s.Verify(ctx, sess.Token); !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected ErrSessionInvalid for expired session, got %v", err)
	}
}
