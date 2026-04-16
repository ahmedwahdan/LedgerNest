package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type userStore interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (model.User, error)
	FindByEmail(ctx context.Context, email string) (repository.UserCredentials, error)
	FindByID(ctx context.Context, id string) (model.User, error)
}

type sessionStore interface {
	Create(ctx context.Context, userID, refreshTokenHash string, expiresAt time.Time) error
	FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (repository.Session, error)
	Rotate(ctx context.Context, currentRefreshTokenHash, nextRefreshTokenHash string, expiresAt time.Time) error
	DeleteByRefreshTokenHash(ctx context.Context, refreshTokenHash string) error
}

// AuthService owns authentication-related business logic.
type AuthService struct {
	users      userStore
	sessions   sessionStore
	tokens     *auth.TokenService
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

type LoginResult struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresAt    time.Time  `json:"expires_at"`
}

func NewAuthService(
	users userStore,
	sessions sessionStore,
	tokens *auth.TokenService,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		sessions:   sessions,
		tokens:     tokens,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (model.User, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, normalizedEmail, string(passwordHash), displayName)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))

	creds, err := s.users.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	now := s.now()

	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(creds.User, s.accessTTL, now)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.sessions.Create(ctx, creds.User.ID, refreshTokenHash, now.Add(s.refreshTTL)); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:         creds.User,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (LoginResult, error) {
	session, err := s.sessions.FindByRefreshTokenHash(ctx, auth.HashToken(strings.TrimSpace(refreshToken)))
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	now := s.now()
	if !session.ExpiresAt.After(now) {
		return LoginResult{}, ErrInvalidCredentials
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(user, s.accessTTL, now)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	nextRefreshToken, nextRefreshTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.sessions.Rotate(ctx, session.RefreshTokenHash, nextRefreshTokenHash, now.Add(s.refreshTTL)); err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	return LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: nextRefreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := auth.HashToken(strings.TrimSpace(refreshToken))

	if err := s.sessions.DeleteByRefreshTokenHash(ctx, hash); err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}

	return nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID string) (model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.User{}, ErrInvalidCredentials
		}
		return model.User{}, err
	}

	return user, nil
}
