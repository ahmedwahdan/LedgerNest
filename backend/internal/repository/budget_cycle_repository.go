package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

var (
	ErrCycleConfigNotFound   = errors.New("cycle config not found")
	ErrCycleSnapshotNotFound = errors.New("cycle snapshot not found")
)

type CreateCycleConfigParams struct {
	HouseholdID   string
	StartDay      int
	EffectiveFrom string // "YYYY-MM-DD"
	CreatedBy     string
}

type CreateSnapshotParams struct {
	HouseholdID string
	CycleStart  string // "YYYY-MM-DD"
	CycleEnd    string
	Label       string
	ConfigID    string
}

type BudgetCycleRepository struct {
	pool *pgxpool.Pool
}

func NewBudgetCycleRepository(pool *pgxpool.Pool) *BudgetCycleRepository {
	return &BudgetCycleRepository{pool: pool}
}

// ── Configs ───────────────────────────────────────────────────────────────────

func (r *BudgetCycleRepository) CreateConfig(ctx context.Context, params CreateCycleConfigParams) (model.BudgetCycleConfig, error) {
	const query = `
		INSERT INTO budget_cycle_configs (household_id, start_day, effective_from, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, household_id, start_day, effective_from::text, created_by, created_at
	`

	var cfg model.BudgetCycleConfig
	err := r.pool.QueryRow(ctx, query,
		params.HouseholdID,
		params.StartDay,
		params.EffectiveFrom,
		params.CreatedBy,
	).Scan(
		&cfg.ID, &cfg.HouseholdID, &cfg.StartDay, &cfg.EffectiveFrom, &cfg.CreatedBy, &cfg.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			// effective_from already set for this household — update it instead.
			return r.updateConfig(ctx, params)
		}
		return model.BudgetCycleConfig{}, fmt.Errorf("create cycle config: %w", err)
	}

	return cfg, nil
}

func (r *BudgetCycleRepository) updateConfig(ctx context.Context, params CreateCycleConfigParams) (model.BudgetCycleConfig, error) {
	const query = `
		UPDATE budget_cycle_configs
		SET start_day = $1
		WHERE household_id = $2 AND effective_from = $3
		RETURNING id, household_id, start_day, effective_from::text, created_by, created_at
	`

	var cfg model.BudgetCycleConfig
	err := r.pool.QueryRow(ctx, query,
		params.StartDay,
		params.HouseholdID,
		params.EffectiveFrom,
	).Scan(
		&cfg.ID, &cfg.HouseholdID, &cfg.StartDay, &cfg.EffectiveFrom, &cfg.CreatedBy, &cfg.CreatedAt,
	)
	if err != nil {
		return model.BudgetCycleConfig{}, fmt.Errorf("update cycle config: %w", err)
	}

	return cfg, nil
}

func (r *BudgetCycleRepository) GetActiveConfig(ctx context.Context, householdID string) (model.BudgetCycleConfig, error) {
	const query = `
		SELECT id, household_id, start_day, effective_from::text, created_by, created_at
		FROM budget_cycle_configs
		WHERE household_id = $1
		ORDER BY effective_from DESC
		LIMIT 1
	`

	var cfg model.BudgetCycleConfig
	err := r.pool.QueryRow(ctx, query, householdID).Scan(
		&cfg.ID, &cfg.HouseholdID, &cfg.StartDay, &cfg.EffectiveFrom, &cfg.CreatedBy, &cfg.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BudgetCycleConfig{}, ErrCycleConfigNotFound
		}
		return model.BudgetCycleConfig{}, fmt.Errorf("get active cycle config: %w", err)
	}

	return cfg, nil
}

// ── Snapshots ─────────────────────────────────────────────────────────────────

func (r *BudgetCycleRepository) CreateSnapshot(ctx context.Context, params CreateSnapshotParams) (model.CycleSnapshot, error) {
	const query = `
		INSERT INTO cycle_snapshots (household_id, cycle_start, cycle_end, label, config_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (household_id, cycle_start, cycle_end) DO UPDATE
		  SET label = EXCLUDED.label, config_id = EXCLUDED.config_id
		RETURNING id, household_id, cycle_start::text, cycle_end::text, label, status::text, config_id, created_at
	`

	var s model.CycleSnapshot
	err := r.pool.QueryRow(ctx, query,
		params.HouseholdID,
		params.CycleStart,
		params.CycleEnd,
		params.Label,
		params.ConfigID,
	).Scan(
		&s.ID, &s.HouseholdID, &s.CycleStart, &s.CycleEnd, &s.Label, &s.Status, &s.ConfigID, &s.CreatedAt,
	)
	if err != nil {
		return model.CycleSnapshot{}, fmt.Errorf("create cycle snapshot: %w", err)
	}

	return s, nil
}

func (r *BudgetCycleRepository) GetOpenSnapshot(ctx context.Context, householdID string) (model.CycleSnapshot, error) {
	const query = `
		SELECT id, household_id, cycle_start::text, cycle_end::text, label, status::text, config_id, created_at
		FROM cycle_snapshots
		WHERE household_id = $1 AND status = 'open'
		ORDER BY cycle_start DESC
		LIMIT 1
	`

	var s model.CycleSnapshot
	err := r.pool.QueryRow(ctx, query, householdID).Scan(
		&s.ID, &s.HouseholdID, &s.CycleStart, &s.CycleEnd, &s.Label, &s.Status, &s.ConfigID, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CycleSnapshot{}, ErrCycleSnapshotNotFound
		}
		return model.CycleSnapshot{}, fmt.Errorf("get open snapshot: %w", err)
	}

	return s, nil
}

func (r *BudgetCycleRepository) GetSnapshotByID(ctx context.Context, snapshotID string) (model.CycleSnapshot, error) {
	const query = `
		SELECT id, household_id, cycle_start::text, cycle_end::text, label, status::text, config_id, created_at
		FROM cycle_snapshots
		WHERE id = $1
	`

	var s model.CycleSnapshot
	err := r.pool.QueryRow(ctx, query, snapshotID).Scan(
		&s.ID, &s.HouseholdID, &s.CycleStart, &s.CycleEnd, &s.Label, &s.Status, &s.ConfigID, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CycleSnapshot{}, ErrCycleSnapshotNotFound
		}
		return model.CycleSnapshot{}, fmt.Errorf("get snapshot by id: %w", err)
	}

	return s, nil
}

func (r *BudgetCycleRepository) ListSnapshots(ctx context.Context, householdID string) ([]model.CycleSnapshot, error) {
	const query = `
		SELECT id, household_id, cycle_start::text, cycle_end::text, label, status::text, config_id, created_at
		FROM cycle_snapshots
		WHERE household_id = $1
		ORDER BY cycle_start DESC
	`

	rows, err := r.pool.Query(ctx, query, householdID)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()

	snapshots := make([]model.CycleSnapshot, 0)
	for rows.Next() {
		var s model.CycleSnapshot
		if err := rows.Scan(
			&s.ID, &s.HouseholdID, &s.CycleStart, &s.CycleEnd, &s.Label, &s.Status, &s.ConfigID, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		snapshots = append(snapshots, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate snapshots: %w", err)
	}

	return snapshots, nil
}

func (r *BudgetCycleRepository) CloseSnapshot(ctx context.Context, snapshotID string) error {
	const query = `UPDATE cycle_snapshots SET status = 'closed' WHERE id = $1`

	if _, err := r.pool.Exec(ctx, query, snapshotID); err != nil {
		return fmt.Errorf("close snapshot: %w", err)
	}

	return nil
}
