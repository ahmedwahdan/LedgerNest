package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

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

func marshalAuditValues(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
