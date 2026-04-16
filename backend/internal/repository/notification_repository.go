package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

type WriteNotificationParams struct {
	UserID   string
	Type     string
	Title    string
	Body     string
	Metadata any
}

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, p WriteNotificationParams) (model.Notification, error) {
	metaJSON, err := json.Marshal(p.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	const query = `
		INSERT INTO notifications (user_id, type, title, body, metadata)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		RETURNING id, user_id, type, title, body, metadata, read_at, created_at
	`

	return scanNotification(r.pool.QueryRow(ctx, query,
		p.UserID, p.Type, p.Title, p.Body, string(metaJSON),
	))
}

func (r *NotificationRepository) List(ctx context.Context, userID string, limit int) ([]model.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	query := fmt.Sprintf(`
		SELECT id, user_id, type, title, body, metadata, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT %d
	`, limit)

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		n, err := scanNotificationRow(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	const query = `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`
	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return count, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, notificationID, userID string) error {
	const query = `
		UPDATE notifications SET read_at = NOW()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, notificationID, userID)
	return err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	const query = `UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// ExistsForBudgetThreshold checks whether a threshold notification was already
// sent for a given budget+snapshot+threshold combination to avoid duplicates.
func (r *NotificationRepository) ExistsForBudgetThreshold(ctx context.Context, userID, budgetID, snapshotID string, threshold int) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM notifications
			WHERE user_id = $1
			  AND type = 'budget_threshold'
			  AND metadata->>'budget_id' = $2
			  AND metadata->>'snapshot_id' = $3
			  AND (metadata->>'threshold')::int = $4
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, userID, budgetID, snapshotID, threshold).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

type scannable interface {
	Scan(dest ...any) error
}

func scanNotification(row scannable) (model.Notification, error) {
	var n model.Notification
	var metaBytes []byte
	var readAt *time.Time

	if err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &metaBytes, &readAt, &n.CreatedAt); err != nil {
		return model.Notification{}, fmt.Errorf("scan notification: %w", err)
	}

	if len(metaBytes) > 0 {
		_ = json.Unmarshal(metaBytes, &n.Metadata)
	}
	n.ReadAt = readAt
	return n, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNotificationRow(row rowScanner) (model.Notification, error) {
	return scanNotification(row)
}
