package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
)

type tokenVerifier interface {
	VerifyAccessToken(token string, now time.Time) (auth.AccessTokenClaims, error)
}

type AuthMiddleware struct {
	tokens tokenVerifier
	now    func() time.Time
}

func NewAuthMiddleware(tokens tokenVerifier) *AuthMiddleware {
	return &AuthMiddleware{
		tokens: tokens,
		now:    time.Now,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerTokenFromHeader(r.Header.Get("Authorization"))
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := m.tokens.VerifyAccessToken(token, m.now())
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid access token")
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.ContextWithAccessTokenClaims(r.Context(), claims)))
	})
}

func bearerTokenFromHeader(header string) (string, bool) {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}

	return token, true
}
