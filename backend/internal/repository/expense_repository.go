package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

var ErrExpenseNotFound = errors.New("expense not found")

type CreateExpenseParams struct {
	UserID             string
	CreatedBy          string
	Amount             string
	Currency           string
	Merchant           string
	CategoryID         *string
	PaymentMethod      string
	Date               string
	Notes              string
	IsRecurring        bool
	RecurrenceInterval *string
}

type UpdateExpenseParams struct {
	ExpenseID          string
	UserID             string
	UpdatedBy          string
	Amount             string
	Currency           string
	Merchant           string
	CategoryID         *string
	PaymentMethod      string
	Date               string
	Notes              string
	IsRecurring        bool
	RecurrenceInterval *string
}

type ListExpenseFilters struct {
	UserID     string
	From       string
	To         string
	Merchant   string
	CategoryID *string
}

type ExpenseRepository struct {
	pool *pgxpool.Pool
}

func NewExpenseRepository(pool *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{pool: pool}
}

func (r *ExpenseRepository) CreatePersonal(ctx context.Context, params CreateExpenseParams) (model.Expense, error) {
	const query = `
		INSERT INTO expenses (
			scope,
			user_id,
			created_by,
			amount,
			currency,
			merchant,
			category_id,
			payment_method,
			date,
			notes,
			is_recurring,
			recurrence_interval
		) VALUES (
			'personal',
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11
		)
		RETURNING
			id,
			scope::text,
			user_id,
			created_by,
			amount::text,
			currency,
			merchant,
			category_id,
			payment_method,
			date::text,
			COALESCE(notes, ''),
			is_recurring,
			recurrence_interval,
			created_at,
			updated_at
	`

	var expense model.Expense
	err := r.pool.QueryRow(
		ctx,
		query,
		params.UserID,
		params.CreatedBy,
		params.Amount,
		params.Currency,
		params.Merchant,
		params.CategoryID,
		params.PaymentMethod,
		params.Date,
		params.Notes,
		params.IsRecurring,
		params.RecurrenceInterval,
	).Scan(
		&expense.ID,
		&expense.Scope,
		&expense.UserID,
		&expense.CreatedBy,
		&expense.Amount,
		&expense.Currency,
		&expense.Merchant,
		&expense.CategoryID,
		&expense.PaymentMethod,
		&expense.Date,
		&expense.Notes,
		&expense.IsRecurring,
		&expense.RecurrenceInterval,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		return model.Expense{}, fmt.Errorf("create personal expense: %w", err)
	}

	return expense, nil
}

func (r *ExpenseRepository) ListPersonalByUserID(ctx context.Context, userID string) ([]model.Expense, error) {
	return r.ListPersonal(ctx, ListExpenseFilters{UserID: userID})
}

func (r *ExpenseRepository) ListPersonal(ctx context.Context, filters ListExpenseFilters) ([]model.Expense, error) {
	baseQuery := `
		SELECT
			id,
			scope::text,
			user_id,
			created_by,
			amount::text,
			currency,
			merchant,
			category_id,
			payment_method,
			date::text,
			COALESCE(notes, ''),
			is_recurring,
			recurrence_interval,
			created_at,
			updated_at
		FROM expenses
		WHERE scope = 'personal'
		  AND is_deleted = FALSE
	`

	conditions := []string{"user_id = $1"}
	args := []any{filters.UserID}
	argIndex := 2

	if filters.From != "" {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argIndex))
		args = append(args, filters.From)
		argIndex++
	}
	if filters.To != "" {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argIndex))
		args = append(args, filters.To)
		argIndex++
	}
	if filters.Merchant != "" {
		conditions = append(conditions, fmt.Sprintf("merchant ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Merchant+"%")
		argIndex++
	}
	if filters.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filters.CategoryID)
		argIndex++
	}

	query := baseQuery + " AND " + strings.Join(conditions, " AND ") + " ORDER BY date DESC, created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list personal expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]model.Expense, 0)
	for rows.Next() {
		var expense model.Expense
		if err := rows.Scan(
			&expense.ID,
			&expense.Scope,
			&expense.UserID,
			&expense.CreatedBy,
			&expense.Amount,
			&expense.Currency,
			&expense.Merchant,
			&expense.CategoryID,
			&expense.PaymentMethod,
			&expense.Date,
			&expense.Notes,
			&expense.IsRecurring,
			&expense.RecurrenceInterval,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan personal expense: %w", err)
		}
		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate personal expenses: %w", err)
	}

	return expenses, nil
}

func (r *ExpenseRepository) GetPersonalByID(ctx context.Context, expenseID, userID string) (model.Expense, error) {
	const query = `
		SELECT
			id,
			scope::text,
			user_id,
			created_by,
			amount::text,
			currency,
			merchant,
			category_id,
			payment_method,
			date::text,
			COALESCE(notes, ''),
			is_recurring,
			recurrence_interval,
			created_at,
			updated_at
		FROM expenses
		WHERE id = $1
		  AND scope = 'personal'
		  AND user_id = $2
		  AND is_deleted = FALSE
	`

	var expense model.Expense
	err := r.pool.QueryRow(ctx, query, expenseID, userID).Scan(
		&expense.ID,
		&expense.Scope,
		&expense.UserID,
		&expense.CreatedBy,
		&expense.Amount,
		&expense.Currency,
		&expense.Merchant,
		&expense.CategoryID,
		&expense.PaymentMethod,
		&expense.Date,
		&expense.Notes,
		&expense.IsRecurring,
		&expense.RecurrenceInterval,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Expense{}, ErrExpenseNotFound
		}
		return model.Expense{}, fmt.Errorf("get personal expense by id: %w", err)
	}

	return expense, nil
}

func (r *ExpenseRepository) UpdatePersonal(ctx context.Context, params UpdateExpenseParams) (model.Expense, error) {
	const query = `
		UPDATE expenses
		SET
			updated_by = $1,
			amount = $2,
			currency = $3,
			merchant = $4,
			category_id = $5,
			payment_method = $6,
			date = $7,
			notes = $8,
			is_recurring = $9,
			recurrence_interval = $10,
			updated_at = NOW()
		WHERE id = $11
		  AND scope = 'personal'
		  AND user_id = $12
		  AND is_deleted = FALSE
		RETURNING
			id,
			scope::text,
			user_id,
			created_by,
			amount::text,
			currency,
			merchant,
			category_id,
			payment_method,
			date::text,
			COALESCE(notes, ''),
			is_recurring,
			recurrence_interval,
			created_at,
			updated_at
	`

	var expense model.Expense
	err := r.pool.QueryRow(
		ctx,
		query,
		params.UpdatedBy,
		params.Amount,
		params.Currency,
		params.Merchant,
		params.CategoryID,
		params.PaymentMethod,
		params.Date,
		params.Notes,
		params.IsRecurring,
		params.RecurrenceInterval,
		params.ExpenseID,
		params.UserID,
	).Scan(
		&expense.ID,
		&expense.Scope,
		&expense.UserID,
		&expense.CreatedBy,
		&expense.Amount,
		&expense.Currency,
		&expense.Merchant,
		&expense.CategoryID,
		&expense.PaymentMethod,
		&expense.Date,
		&expense.Notes,
		&expense.IsRecurring,
		&expense.RecurrenceInterval,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Expense{}, ErrExpenseNotFound
		}
		return model.Expense{}, fmt.Errorf("update personal expense: %w", err)
	}

	return expense, nil
}

func (r *ExpenseRepository) DeletePersonal(ctx context.Context, expenseID, userID, deletedBy string) error {
	const query = `
		UPDATE expenses
		SET
			is_deleted = TRUE,
			deleted_at = NOW(),
			deleted_by = $1,
			updated_by = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND scope = 'personal'
		  AND user_id = $3
		  AND is_deleted = FALSE
	`

	tag, err := r.pool.Exec(ctx, query, deletedBy, expenseID, userID)
	if err != nil {
		return fmt.Errorf("delete personal expense: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrExpenseNotFound
	}

	return nil
}
