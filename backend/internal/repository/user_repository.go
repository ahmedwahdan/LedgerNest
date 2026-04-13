package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

var ErrUserEmailExists = errors.New("user email already exists")
var ErrUserNotFound = errors.New("user not found")

type UserCredentials struct {
	User         model.User
	PasswordHash string
}

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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (UserCredentials, error) {
	const query = `
		SELECT id, email, password_hash, display_name, preferred_currency, verified_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var creds UserCredentials

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&creds.User.ID,
		&creds.User.Email,
		&creds.PasswordHash,
		&creds.User.DisplayName,
		&creds.User.PreferredCurrency,
		&creds.User.VerifiedAt,
		&creds.User.CreatedAt,
		&creds.User.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserCredentials{}, ErrUserNotFound
		}
		return UserCredentials{}, fmt.Errorf("find user by email: %w", err)
	}

	return creds, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (model.User, error) {
	const query = `
		SELECT id, email, display_name, preferred_currency, verified_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PreferredCurrency,
		&user.VerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}
