package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

func TestAuthServiceLogin(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("super-secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	now := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)
	users := &stubUserStore{
		creds: repository.UserCredentials{
			User: model.User{
				ID:                "user-1",
				Email:             "user@example.com",
				DisplayName:       "User",
				PreferredCurrency: "USD",
			},
			PasswordHash: string(passwordHash),
		},
	}
	sessions := &stubSessionStore{}
	service := NewAuthService(
		users,
		sessions,
		auth.NewTokenService("12345678901234567890123456789012"),
		15*time.Minute,
		24*time.Hour,
	)
	service.now = func() time.Time { return now }

	result, err := service.Login(context.Background(), "User@Example.com", "super-secret")
	if err != nil {
		t.Fatalf("login returned error: %v", err)
	}

	if result.User.Email != "user@example.com" {
		t.Fatalf("unexpected email: %s", result.User.Email)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	if result.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %s", result.TokenType)
	}
	if !result.ExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("unexpected access expiry: %s", result.ExpiresAt)
	}
	if sessions.userID != "user-1" {
		t.Fatalf("unexpected session user id: %s", sessions.userID)
	}
	if sessions.refreshTokenHash == "" {
		t.Fatal("expected refresh token hash to be stored")
	}
	if !sessions.expiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("unexpected refresh expiry: %s", sessions.expiresAt)
	}
	if users.lookupEmail != "user@example.com" {
		t.Fatalf("expected normalized email lookup, got %s", users.lookupEmail)
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("super-secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := NewAuthService(
		&stubUserStore{
			creds: repository.UserCredentials{
				User:         model.User{ID: "user-1", Email: "user@example.com"},
				PasswordHash: string(passwordHash),
			},
		},
		&stubSessionStore{},
		auth.NewTokenService("12345678901234567890123456789012"),
		15*time.Minute,
		24*time.Hour,
	)

	_, err = service.Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestAuthServiceRefresh(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 13, 16, 0, 0, 0, time.UTC)
	users := &stubUserStore{
		user: model.User{
			ID:                "user-1",
			Email:             "user@example.com",
			DisplayName:       "User",
			PreferredCurrency: "USD",
		},
	}
	sessions := &stubSessionStore{
		session: repository.Session{
			ID:               "session-1",
			UserID:           "user-1",
			RefreshTokenHash: auth.HashToken("refresh-token"),
			ExpiresAt:        now.Add(time.Hour),
		},
	}
	service := NewAuthService(
		users,
		sessions,
		auth.NewTokenService("12345678901234567890123456789012"),
		15*time.Minute,
		24*time.Hour,
	)
	service.now = func() time.Time { return now }

	result, err := service.Refresh(context.Background(), "refresh-token")
	if err != nil {
		t.Fatalf("refresh returned error: %v", err)
	}

	if result.User.ID != "user-1" {
		t.Fatalf("unexpected user id: %s", result.User.ID)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatal("expected both tokens")
	}
	if sessions.findRefreshTokenHash != auth.HashToken("refresh-token") {
		t.Fatalf("unexpected refresh token hash lookup: %s", sessions.findRefreshTokenHash)
	}
	if sessions.rotatedCurrentHash != auth.HashToken("refresh-token") {
		t.Fatalf("unexpected rotate current hash: %s", sessions.rotatedCurrentHash)
	}
	if sessions.rotatedNextHash == "" {
		t.Fatal("expected rotated next hash")
	}
}

func TestAuthServiceCurrentUser(t *testing.T) {
	t.Parallel()

	service := NewAuthService(
		&stubUserStore{
			user: model.User{ID: "user-1", Email: "user@example.com"},
		},
		&stubSessionStore{},
		auth.NewTokenService("12345678901234567890123456789012"),
		15*time.Minute,
		24*time.Hour,
	)

	user, err := service.CurrentUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("current user returned error: %v", err)
	}
	if user.ID != "user-1" {
		t.Fatalf("unexpected user id: %s", user.ID)
	}
}

type stubUserStore struct {
	creds       repository.UserCredentials
	err         error
	lookupEmail string
	user        model.User
	findByIDID  string
}

func (s *stubUserStore) Create(context.Context, string, string, string) (model.User, error) {
	return model.User{}, nil
}

func (s *stubUserStore) FindByEmail(_ context.Context, email string) (repository.UserCredentials, error) {
	s.lookupEmail = email
	if s.err != nil {
		return repository.UserCredentials{}, s.err
	}
	return s.creds, nil
}

func (s *stubUserStore) FindByID(_ context.Context, id string) (model.User, error) {
	s.findByIDID = id
	if s.err != nil {
		return model.User{}, s.err
	}
	return s.user, nil
}

type stubSessionStore struct {
	userID               string
	refreshTokenHash     string
	expiresAt            time.Time
	session              repository.Session
	findRefreshTokenHash string
	rotatedCurrentHash   string
	rotatedNextHash      string
	rotatedExpiresAt     time.Time
}

func (s *stubSessionStore) Create(_ context.Context, userID, refreshTokenHash string, expiresAt time.Time) error {
	s.userID = userID
	s.refreshTokenHash = refreshTokenHash
	s.expiresAt = expiresAt
	return nil
}

func (s *stubSessionStore) FindByRefreshTokenHash(_ context.Context, refreshTokenHash string) (repository.Session, error) {
	s.findRefreshTokenHash = refreshTokenHash
	if s.session.ID == "" {
		return repository.Session{}, repository.ErrSessionNotFound
	}
	return s.session, nil
}

func (s *stubSessionStore) Rotate(_ context.Context, currentRefreshTokenHash, nextRefreshTokenHash string, expiresAt time.Time) error {
	s.rotatedCurrentHash = currentRefreshTokenHash
	s.rotatedNextHash = nextRefreshTokenHash
	s.rotatedExpiresAt = expiresAt
	return nil
}

func (s *stubSessionStore) DeleteByRefreshTokenHash(context.Context, string) error {
	return nil
}
