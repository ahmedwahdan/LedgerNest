package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
)

func TestAuthMiddlewareRequireAuth(t *testing.T) {
	t.Parallel()

	middleware := NewAuthMiddleware(&stubTokenVerifier{
		claims: auth.AccessTokenClaims{UserID: "user-1", Email: "user@example.com"},
	})

	handlerCalled := false
	next := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
		if !ok || claims.UserID != "user-1" {
			t.Fatalf("expected claims in context, got %#v", claims)
		}
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/auth/me", http.NoBody)
	request.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()

	next.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !handlerCalled {
		t.Fatal("expected next handler to be called")
	}
}

func TestAuthMiddlewareRequireAuthMissingToken(t *testing.T) {
	t.Parallel()

	middleware := NewAuthMiddleware(&stubTokenVerifier{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/me", http.NoBody)

	middleware.RequireAuth(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

type stubTokenVerifier struct {
	claims auth.AccessTokenClaims
	err    error
}

func (s *stubTokenVerifier) VerifyAccessToken(string, time.Time) (auth.AccessTokenClaims, error) {
	if s.err != nil {
		return auth.AccessTokenClaims{}, s.err
	}
	return s.claims, nil
}
