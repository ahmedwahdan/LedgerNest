package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

func TestAuthHandlerLogin(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{
		loginResult: service.LoginResult{
			User:         model.User{ID: "user-1", Email: "user@example.com"},
			AccessToken:  "access",
			RefreshToken: "refresh",
			TokenType:    "Bearer",
			ExpiresAt:    time.Date(2026, 4, 13, 16, 15, 0, 0, time.UTC),
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"super-secret"}`))

	handler.Login(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response service.LoginResult
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.AccessToken != "access" {
		t.Fatalf("unexpected access token: %s", response.AccessToken)
	}
}

func TestAuthHandlerLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{loginErr: service.ErrInvalidCredentials})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"wrong"}`))

	handler.Login(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestAuthHandlerRefresh(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{
		refreshResult: service.LoginResult{
			User:         model.User{ID: "user-1", Email: "user@example.com"},
			AccessToken:  "next-access",
			RefreshToken: "next-refresh",
			TokenType:    "Bearer",
			ExpiresAt:    time.Date(2026, 4, 13, 16, 15, 0, 0, time.UTC),
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"refresh-token"}`))

	handler.Refresh(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestAuthHandlerMe(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{
		currentUser: model.User{ID: "user-1", Email: "user@example.com"},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))

	handler.Me(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

type stubAuthService struct {
	loginResult    service.LoginResult
	loginErr       error
	refreshResult  service.LoginResult
	refreshErr     error
	currentUser    model.User
	currentUserErr error
}

func (s *stubAuthService) Register(context.Context, string, string, string) (model.User, error) {
	return model.User{}, nil
}

func (s *stubAuthService) Login(context.Context, string, string) (service.LoginResult, error) {
	if s.loginErr != nil {
		return service.LoginResult{}, s.loginErr
	}
	return s.loginResult, nil
}

func (s *stubAuthService) Refresh(context.Context, string) (service.LoginResult, error) {
	if s.refreshErr != nil {
		return service.LoginResult{}, s.refreshErr
	}
	return s.refreshResult, nil
}

func (s *stubAuthService) CurrentUser(context.Context, string) (model.User, error) {
	if s.currentUserErr != nil {
		return model.User{}, s.currentUserErr
	}
	return s.currentUser, nil
}
