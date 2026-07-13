// Package accounts is the user-account + session domain for the bike shop. Accounts are its own
// service (cmd/accounts) — the sole authority on identity — persisted via internal/awskv, same as
// every other self-service store in this app. There is no JWT/shared-signing-secret here: a session
// is just an opaque random token the accounts service looks up itself, verified over a
// mutual-auth'd east-west call rather than trusted blind.
package accounts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/asanexample/alpha-shop/internal/awskv"
	"golang.org/x/crypto/bcrypt"
)

// ErrEmailTaken is returned by Signup when the (normalized) email already has an account.
var ErrEmailTaken = errors.New("email already registered")

// ErrInvalidCredentials is returned by Login on a wrong email or password (deliberately the same
// error for both — don't reveal whether an email is registered).
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrSessionInvalid is returned by Verify for a missing, unknown, or expired token.
var ErrSessionInvalid = errors.New("invalid or expired session")

const sessionTTL = 7 * 24 * time.Hour

// User is one registered account. ID is the normalized (lowercased, trimmed) email — the DynamoDB
// key, so login-by-email is a direct Get with no secondary index needed.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Session is one active login. Token is the DynamoDB key — a crypto/rand opaque string, same idiom
// as orders' orderID().
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Store persists users and sessions in two separate awskv tables.
type Store struct {
	users    awskv.Store
	sessions awskv.Store
	now      func() time.Time
}

// New returns an accounts Store over the given users/sessions key-value backends.
func New(users, sessions awskv.Store) *Store {
	return &Store{users: users, sessions: sessions, now: func() time.Time { return time.Now().UTC() }}
}

// Backend reports the underlying kv backends (for startup logging).
func (s *Store) Backend() string { return s.users.Backend() + "/" + s.sessions.Backend() }

// Signup creates a new account (409-equivalent ErrEmailTaken if the email is already registered)
// and immediately mints a session, same as Login — the caller doesn't need a second round-trip.
func (s *Store) Signup(ctx context.Context, email, password, name string) (User, Session, error) {
	id := normalizeEmail(email)
	if _, found, err := s.users.Get(ctx, id); err != nil {
		return User{}, Session{}, err
	} else if found {
		return User{}, Session{}, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, Session{}, err
	}
	u := User{ID: id, Email: id, Name: name, PasswordHash: string(hash), CreatedAt: s.now()}
	if err := s.saveUser(ctx, u); err != nil {
		return User{}, Session{}, err
	}

	sess, err := s.newSession(ctx, u.ID)
	return u, sess, err
}

// Login verifies the password and mints a new session on success.
func (s *Store) Login(ctx context.Context, email, password string) (User, Session, error) {
	id := normalizeEmail(email)
	doc, found, err := s.users.Get(ctx, id)
	if err != nil {
		return User{}, Session{}, err
	}
	if !found {
		return User{}, Session{}, ErrInvalidCredentials
	}
	var u User
	if err := json.Unmarshal(doc, &u); err != nil {
		return User{}, Session{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return User{}, Session{}, ErrInvalidCredentials
	}
	sess, err := s.newSession(ctx, u.ID)
	return u, sess, err
}

// Verify looks up a session token and returns its user (ErrSessionInvalid if missing/expired).
func (s *Store) Verify(ctx context.Context, token string) (User, error) {
	doc, found, err := s.sessions.Get(ctx, token)
	if err != nil {
		return User{}, err
	}
	if !found {
		return User{}, ErrSessionInvalid
	}
	var sess Session
	if err := json.Unmarshal(doc, &sess); err != nil {
		return User{}, err
	}
	if s.now().After(sess.ExpiresAt) {
		return User{}, ErrSessionInvalid
	}

	udoc, found, err := s.users.Get(ctx, sess.UserID)
	if err != nil {
		return User{}, err
	}
	if !found {
		return User{}, ErrSessionInvalid
	}
	var u User
	if err := json.Unmarshal(udoc, &u); err != nil {
		return User{}, err
	}
	return u, nil
}

// Logout deletes a session token. Deleting an unknown token is a no-op, not an error.
func (s *Store) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}

func (s *Store) newSession(ctx context.Context, userID string) (Session, error) {
	token, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	sess := Session{Token: token, UserID: userID, CreatedAt: s.now(), ExpiresAt: s.now().Add(sessionTTL)}
	b, err := json.Marshal(sess)
	if err != nil {
		return Session{}, err
	}
	if err := s.sessions.Put(ctx, token, b); err != nil {
		return Session{}, err
	}
	return sess, nil
}

func (s *Store) saveUser(ctx context.Context, u User) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return s.users.Put(ctx, u.ID, b)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
