package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

type ActivityFilters struct {
	UserID     *string
	EntityType *string
	From       *time.Time
	To         *time.Time
	Page       int // 1-indexed
	Limit      int // max 100
}

type WriteAuditParams struct {
	UserID     *string
	Action     string // create | update | delete | restore
	EntityType string
	EntityID   string
	OldValues  any
	NewValues  any
	IPAddress  *string
}

type AuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

func (r *AuditLogRepository) Write(ctx context.Context, params WriteAuditParams) error {
	oldJSON, err := marshalAuditValues(params.OldValues)
	if err != nil {
		return fmt.Errorf("marshal old_values: %w", err)
	}

	newJSON, err := marshalAuditValues(params.NewValues)
	if err != nil {
		return fmt.Errorf("marshal new_values: %w", err)
	}

	const query = `
		INSERT INTO audit_log (user_id, action, entity_type, entity_id, old_values, new_values, ip_address)
		VALUES ($1, $2::audit_action, $3, $4, $5, $6, $7::inet)
	`

	if _, err := r.pool.Exec(ctx, query,
		params.UserID,
		params.Action,
		params.EntityType,
		params.EntityID,
		oldJSON,
		newJSON,
		params.IPAddress,
	); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}

func (r *AuditLogRepository) ListByEntity(ctx context.Context, entityType, entityID string) ([]model.AuditLogEntry, error) {
	const query = `
		SELECT id, user_id, action::text, entity_type, entity_id::text,
		       old_values, new_values, ip_address::text, created_at
		FROM audit_log
		WHERE entity_type = $1 AND entity_id = $2::uuid
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("list audit log: %w", err)
	}
	defer rows.Close()

	entries := make([]model.AuditLogEntry, 0)
	for rows.Next() {
		var e model.AuditLogEntry
		var oldRaw, newRaw []byte
		var ipStr *string

		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Action, &e.EntityType, &e.EntityID,
			&oldRaw, &newRaw, &ipStr, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}

		e.IPAddress = ipStr

		if len(oldRaw) > 0 && string(oldRaw) != "null" {
			if err := json.Unmarshal(oldRaw, &e.OldValues); err != nil {
				return nil, fmt.Errorf("unmarshal old_values: %w", err)
			}
		}
		if len(newRaw) > 0 && string(newRaw) != "null" {
			if err := json.Unmarshal(newRaw, &e.NewValues); err != nil {
				return nil, fmt.Errorf("unmarshal new_values: %w", err)
			}
		}

		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit log: %w", err)
	}

	return entries, nil
}

// ListActivity returns paginated audit log entries matching the given filters.
func (r *AuditLogRepository) ListActivity(ctx context.Context, f ActivityFilters) ([]model.AuditLogEntry, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	var conds []string
	var args []any
	argN := 1

	if f.UserID != nil {
		conds = append(conds, fmt.Sprintf("user_id = $%d::uuid", argN))
		args = append(args, *f.UserID)
		argN++
	}
	if f.EntityType != nil {
		conds = append(conds, fmt.Sprintf("entity_type = $%d", argN))
		args = append(args, *f.EntityType)
		argN++
	}
	if f.From != nil {
		conds = append(conds, fmt.Sprintf("created_at >= $%d", argN))
		args = append(args, *f.From)
		argN++
	}
	if f.To != nil {
		conds = append(conds, fmt.Sprintf("created_at <= $%d", argN))
		args = append(args, *f.To)
		argN++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	args = append(args, f.Limit, offset)
	query := fmt.Sprintf(`
		SELECT id, user_id, action::text, entity_type, entity_id::text,
		       old_values, new_values, ip_address::text, created_at
		FROM audit_log
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argN, argN+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list activity: %w", err)
	}
	defer rows.Close()

	entries := make([]model.AuditLogEntry, 0)
	for rows.Next() {
		var e model.AuditLogEntry
		var oldRaw, newRaw []byte
		var ipStr *string

		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Action, &e.EntityType, &e.EntityID,
			&oldRaw, &newRaw, &ipStr, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan activity row: %w", err)
		}

		e.IPAddress = ipStr

		if len(oldRaw) > 0 && string(oldRaw) != "null" {
			if err := json.Unmarshal(oldRaw, &e.OldValues); err != nil {
				return nil, fmt.Errorf("unmarshal old_values: %w", err)
			}
		}
		if len(newRaw) > 0 && string(newRaw) != "null" {
			if err := json.Unmarshal(newRaw, &e.NewValues); err != nil {
				return nil, fmt.Errorf("unmarshal new_values: %w", err)
			}
		}

		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activity: %w", err)
	}

	return entries, nil
}

func marshalAuditValues(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
