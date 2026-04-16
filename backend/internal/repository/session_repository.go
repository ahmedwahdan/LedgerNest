package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

// SessionRepository persists refresh-token backed user sessions.
type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, userID, refreshTokenHash string, expiresAt time.Time) error {
	const query = `
		INSERT INTO user_sessions (
			user_id,
			refresh_token_hash,
			expires_at
		) VALUES ($1, $2, $3)
	`

	if _, err := r.pool.Exec(ctx, query, userID, refreshTokenHash, expiresAt); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (r *SessionRepository) FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (Session, error) {
	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, created_at
		FROM user_sessions
		WHERE refresh_token_hash = $1
	`

	var session Session
	err := r.pool.QueryRow(ctx, query, refreshTokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("find session by refresh token hash: %w", err)
	}

	return session, nil
}

func (r *SessionRepository) Rotate(ctx context.Context, currentRefreshTokenHash, nextRefreshTokenHash string, expiresAt time.Time) error {
	const query = `
		UPDATE user_sessions
		SET refresh_token_hash = $1, expires_at = $2
		WHERE refresh_token_hash = $3
	`

	tag, err := r.pool.Exec(ctx, query, nextRefreshTokenHash, expiresAt, currentRefreshTokenHash)
	if err != nil {
		return fmt.Errorf("rotate session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepository) DeleteByRefreshTokenHash(ctx context.Context, refreshTokenHash string) error {
	const query = `DELETE FROM user_sessions WHERE refresh_token_hash = $1`

	tag, err := r.pool.Exec(ctx, query, refreshTokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}
