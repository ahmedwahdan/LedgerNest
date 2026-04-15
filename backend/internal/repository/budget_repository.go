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
	ErrBudgetNotFound = errors.New("budget not found")
	ErrBudgetConflict = errors.New("a budget for this category already exists in this snapshot")
)

type CreateBudgetParams struct {
	Scope           string
	UserID          *string
	HouseholdID     *string
	CategoryID      *string
	CycleSnapshotID string
	Amount          string
}

type UpdateBudgetParams struct {
	BudgetID string
	// RequesterID and scope ownership must be validated by the service before calling.
	Amount string
}

type ListBudgetFilters struct {
	CycleSnapshotID string
	Scope           *string
	UserID          *string
	HouseholdID     *string
}

type BudgetHealthRow struct {
	BudgetID     string
	CategoryID   *string
	CategoryName *string
	Amount       string
	Rollover     string
	Spent        string
}

type BudgetRepository struct {
	pool *pgxpool.Pool
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{pool: pool}
}

func (r *BudgetRepository) Create(ctx context.Context, params CreateBudgetParams) (model.Budget, error) {
	const query = `
		INSERT INTO budgets (scope, user_id, household_id, category_id, cycle_snapshot_id, amount)
		VALUES ($1::budget_scope, $2, $3, $4, $5, $6)
		RETURNING id, scope::text, user_id, household_id, category_id,
		          cycle_snapshot_id, amount::text, rollover_amount::text, created_at, updated_at
	`

	var b model.Budget
	err := r.pool.QueryRow(ctx, query,
		params.Scope,
		params.UserID,
		params.HouseholdID,
		params.CategoryID,
		params.CycleSnapshotID,
		params.Amount,
	).Scan(
		&b.ID, &b.Scope, &b.UserID, &b.HouseholdID, &b.CategoryID,
		&b.CycleSnapshotID, &b.Amount, &b.RolloverAmount, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Budget{}, ErrBudgetConflict
		}
		return model.Budget{}, fmt.Errorf("create budget: %w", err)
	}

	return b, nil
}

func (r *BudgetRepository) List(ctx context.Context, filters ListBudgetFilters) ([]model.Budget, error) {
	// Build query dynamically based on provided filters.
	query := `
		SELECT id, scope::text, user_id, household_id, category_id,
		       cycle_snapshot_id, amount::text, rollover_amount::text, created_at, updated_at
		FROM budgets
		WHERE cycle_snapshot_id = $1
	`
	args := []any{filters.CycleSnapshotID}
	idx := 2

	if filters.Scope != nil {
		query += fmt.Sprintf(" AND scope = $%d::budget_scope", idx)
		args = append(args, *filters.Scope)
		idx++
	}
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *filters.UserID)
		idx++
	}
	if filters.HouseholdID != nil {
		query += fmt.Sprintf(" AND household_id = $%d", idx)
		args = append(args, *filters.HouseholdID)
		idx++
	}
	query += " ORDER BY created_at ASC"
	_ = idx

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	budgets := make([]model.Budget, 0)
	for rows.Next() {
		var b model.Budget
		if err := rows.Scan(
			&b.ID, &b.Scope, &b.UserID, &b.HouseholdID, &b.CategoryID,
			&b.CycleSnapshotID, &b.Amount, &b.RolloverAmount, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		budgets = append(budgets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate budgets: %w", err)
	}

	return budgets, nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, budgetID string) (model.Budget, error) {
	const query = `
		SELECT id, scope::text, user_id, household_id, category_id,
		       cycle_snapshot_id, amount::text, rollover_amount::text, created_at, updated_at
		FROM budgets
		WHERE id = $1
	`

	var b model.Budget
	err := r.pool.QueryRow(ctx, query, budgetID).Scan(
		&b.ID, &b.Scope, &b.UserID, &b.HouseholdID, &b.CategoryID,
		&b.CycleSnapshotID, &b.Amount, &b.RolloverAmount, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Budget{}, ErrBudgetNotFound
		}
		return model.Budget{}, fmt.Errorf("get budget: %w", err)
	}

	return b, nil
}

func (r *BudgetRepository) Update(ctx context.Context, params UpdateBudgetParams) (model.Budget, error) {
	const query = `
		UPDATE budgets SET amount = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, scope::text, user_id, household_id, category_id,
		          cycle_snapshot_id, amount::text, rollover_amount::text, created_at, updated_at
	`

	var b model.Budget
	err := r.pool.QueryRow(ctx, query, params.Amount, params.BudgetID).Scan(
		&b.ID, &b.Scope, &b.UserID, &b.HouseholdID, &b.CategoryID,
		&b.CycleSnapshotID, &b.Amount, &b.RolloverAmount, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Budget{}, ErrBudgetNotFound
		}
		return model.Budget{}, fmt.Errorf("update budget: %w", err)
	}

	return b, nil
}

func (r *BudgetRepository) Delete(ctx context.Context, budgetID string) error {
	const query = `DELETE FROM budgets WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, budgetID)
	if err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBudgetNotFound
	}

	return nil
}

// GetHealthData returns budgets for a snapshot joined with the amount spent during
// the snapshot period. It handles both category-specific and overall-cap (NULL category) budgets.
func (r *BudgetRepository) GetHealthData(ctx context.Context, snapshotID, scope string, userID, householdID *string) ([]BudgetHealthRow, error) {
	// Uses a recursive CTE so that a budget set at a parent category (e.g. "Food & Dining")
	// automatically aggregates spending from all descendant categories (Groceries, Restaurants, …).
	const query = `
		WITH RECURSIVE snap AS (
			SELECT cycle_start, cycle_end FROM cycle_snapshots WHERE id = $1
		),
		budget_roots AS (
			SELECT id AS budget_id, category_id AS cat_id
			FROM budgets
			WHERE cycle_snapshot_id = $1
			  AND scope = $2::budget_scope
			  AND ($3::uuid IS NULL OR user_id = $3)
			  AND ($4::uuid IS NULL OR household_id = $4)
			  AND category_id IS NOT NULL
		),
		cat_tree(budget_id, cat_id) AS (
			SELECT budget_id, cat_id FROM budget_roots
			UNION ALL
			SELECT ct.budget_id, c.id
			FROM cat_tree ct
			JOIN categories c ON c.parent_id = ct.cat_id
		),
		spent_by_budget AS (
			SELECT ct.budget_id, SUM(e.amount) AS total
			FROM cat_tree ct
			JOIN expenses e ON e.category_id = ct.cat_id
			CROSS JOIN snap
			WHERE e.scope = $2::budget_scope
			  AND ($3::uuid IS NULL OR e.user_id = $3)
			  AND ($4::uuid IS NULL OR e.household_id = $4)
			  AND e.date BETWEEN snap.cycle_start AND snap.cycle_end
			  AND e.is_deleted = FALSE
			GROUP BY ct.budget_id
		),
		total_spent AS (
			SELECT COALESCE(SUM(e.amount), 0) AS grand_total
			FROM expenses e
			CROSS JOIN snap
			WHERE e.scope = $2::budget_scope
			  AND ($3::uuid IS NULL OR e.user_id = $3)
			  AND ($4::uuid IS NULL OR e.household_id = $4)
			  AND e.date BETWEEN snap.cycle_start AND snap.cycle_end
			  AND e.is_deleted = FALSE
		)
		SELECT
			b.id,
			b.category_id,
			c.name AS category_name,
			b.amount::text,
			b.rollover_amount::text,
			COALESCE(
				CASE
					WHEN b.category_id IS NULL THEN (SELECT grand_total FROM total_spent)
					ELSE sbb.total
				END,
				0
			)::text AS spent
		FROM budgets b
		LEFT JOIN categories c ON c.id = b.category_id
		LEFT JOIN spent_by_budget sbb ON sbb.budget_id = b.id
		WHERE b.cycle_snapshot_id = $1
		  AND b.scope = $2::budget_scope
		  AND ($3::uuid IS NULL OR b.user_id = $3)
		  AND ($4::uuid IS NULL OR b.household_id = $4)
		ORDER BY b.category_id NULLS LAST, c.name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID, scope, userID, householdID)
	if err != nil {
		return nil, fmt.Errorf("get health data: %w", err)
	}
	defer rows.Close()

	result := make([]BudgetHealthRow, 0)
	for rows.Next() {
		var row BudgetHealthRow
		if err := rows.Scan(
			&row.BudgetID,
			&row.CategoryID,
			&row.CategoryName,
			&row.Amount,
			&row.Rollover,
			&row.Spent,
		); err != nil {
			return nil, fmt.Errorf("scan health row: %w", err)
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate health rows: %w", err)
	}

	return result, nil
}
