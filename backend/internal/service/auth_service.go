package service

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

// AuthService owns authentication-related business logic.
type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
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
