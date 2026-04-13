package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

type TokenService struct {
	secret []byte
}

var ErrInvalidToken = errors.New("invalid token")

type AccessTokenClaims struct {
	UserID string
	Email  string
}

type jwtClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Issued  int64  `json:"iat"`
	Expiry  int64  `json:"exp"`
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: []byte(secret)}
}

func (s *TokenService) GenerateAccessToken(user model.User, ttl time.Duration, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(ttl).UTC()

	header, err := encodeJWTPart(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode jwt header: %w", err)
	}

	claims, err := encodeJWTPart(map[string]any{
		"sub":   user.ID,
		"email": user.Email,
		"iat":   now.UTC().Unix(),
		"exp":   expiresAt.Unix(),
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode jwt claims: %w", err)
	}

	unsigned := header + "." + claims
	signature := s.sign(unsigned)

	return unsigned + "." + signature, expiresAt, nil
}

func (s *TokenService) VerifyAccessToken(token string, now time.Time) (AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := s.sign(unsigned)
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedSignature)) != 1 {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	payload, err := decodeJWTPart(parts[1])
	if err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	if claims.Subject == "" || claims.Email == "" || claims.Expiry == 0 {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	if now.UTC().Unix() >= claims.Expiry {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	return AccessTokenClaims{
		UserID: claims.Subject,
		Email:  claims.Email,
	}, nil
}

func (s *TokenService) GenerateRefreshToken() (plainToken string, tokenHash string, err error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	plainToken = base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHashBytes := sha256.Sum256([]byte(plainToken))

	return plainToken, hex.EncodeToString(tokenHashBytes[:]), nil
}

func encodeJWTPart(payload any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(body), nil
}

func HashToken(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func decodeJWTPart(part string) ([]byte, error) {
	body, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *TokenService) sign(unsignedToken string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(unsignedToken))

	return strings.TrimRight(base64.URLEncoding.EncodeToString(mac.Sum(nil)), "=")
}
