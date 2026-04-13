package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryNameConflict = errors.New("category name already exists")

type CreateCategoryParams struct {
	HouseholdID string
	Name        string
	ParentID    *string
	Icon        *string
	Color       *string
}

type UpdateCategoryParams struct {
	CategoryID  string
	HouseholdID string
	Name        string
	ParentID    *string
	Icon        *string
	Color       *string
}

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) List(ctx context.Context, householdID *string) ([]model.Category, error) {
	const query = `
		SELECT id, household_id, name, parent_id, icon, color, is_system, created_at
		FROM categories
		WHERE is_system = TRUE
		   OR household_id = $1
		ORDER BY is_system DESC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, householdID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]model.Category, 0)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(
			&c.ID,
			&c.HouseholdID,
			&c.Name,
			&c.ParentID,
			&c.Icon,
			&c.Color,
			&c.IsSystem,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, categoryID string) (model.Category, error) {
	const query = `
		SELECT id, household_id, name, parent_id, icon, color, is_system, created_at
		FROM categories
		WHERE id = $1
	`

	var c model.Category
	err := r.pool.QueryRow(ctx, query, categoryID).Scan(
		&c.ID,
		&c.HouseholdID,
		&c.Name,
		&c.ParentID,
		&c.Icon,
		&c.Color,
		&c.IsSystem,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, ErrCategoryNotFound
		}
		return model.Category{}, fmt.Errorf("get category by id: %w", err)
	}

	return c, nil
}

func (r *CategoryRepository) Create(ctx context.Context, params CreateCategoryParams) (model.Category, error) {
	const query = `
		INSERT INTO categories (household_id, name, parent_id, icon, color)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, household_id, name, parent_id, icon, color, is_system, created_at
	`

	var c model.Category
	err := r.pool.QueryRow(ctx, query,
		params.HouseholdID,
		params.Name,
		params.ParentID,
		params.Icon,
		params.Color,
	).Scan(
		&c.ID,
		&c.HouseholdID,
		&c.Name,
		&c.ParentID,
		&c.Icon,
		&c.Color,
		&c.IsSystem,
		&c.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Category{}, ErrCategoryNameConflict
		}
		return model.Category{}, fmt.Errorf("create category: %w", err)
	}

	return c, nil
}

func (r *CategoryRepository) Update(ctx context.Context, params UpdateCategoryParams) (model.Category, error) {
	const query = `
		UPDATE categories
		SET name = $1, parent_id = $2, icon = $3, color = $4
		WHERE id = $5
		  AND household_id = $6
		  AND is_system = FALSE
		RETURNING id, household_id, name, parent_id, icon, color, is_system, created_at
	`

	var c model.Category
	err := r.pool.QueryRow(ctx, query,
		params.Name,
		params.ParentID,
		params.Icon,
		params.Color,
		params.CategoryID,
		params.HouseholdID,
	).Scan(
		&c.ID,
		&c.HouseholdID,
		&c.Name,
		&c.ParentID,
		&c.Icon,
		&c.Color,
		&c.IsSystem,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, ErrCategoryNotFound
		}
		if isUniqueViolation(err) {
			return model.Category{}, ErrCategoryNameConflict
		}
		return model.Category{}, fmt.Errorf("update category: %w", err)
	}

	return c, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryID, householdID string) error {
	const query = `
		DELETE FROM categories
		WHERE id = $1
		  AND household_id = $2
		  AND is_system = FALSE
	`

	tag, err := r.pool.Exec(ctx, query, categoryID, householdID)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
