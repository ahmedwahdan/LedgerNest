package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

var ErrUserEmailExists = errors.New("user email already exists")

// UserRepository persists and reads users from Postgres.
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, displayName string) (model.User, error) {
	const query = `
		INSERT INTO users (
			email,
			password_hash,
			display_name
		) VALUES ($1, $2, $3)
		RETURNING id, email, display_name, preferred_currency, verified_at, created_at, updated_at
	`

	var user model.User

	err := r.pool.QueryRow(ctx, query, email, passwordHash, displayName).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PreferredCurrency,
		&user.VerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrUserEmailExists
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
