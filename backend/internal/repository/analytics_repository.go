package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ─── param / result types ────────────────────────────────────────────────────

type AnalyticsFilters struct {
	// Exactly one of UserID / HouseholdID is set.
	UserID      *string
	HouseholdID *string
	From        time.Time
	To          time.Time
}

type SpendingByCategory struct {
	CategoryID   *string `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Total        string  `json:"total"`
	Count        int     `json:"count"`
	PctOfTotal   float64 `json:"pct_of_total"`
}

type MonthlyTrend struct {
	Month string `json:"month"` // "2025-03"
	Total string `json:"total"`
	Count int    `json:"count"`
}

type TopMerchant struct {
	Merchant string `json:"merchant"`
	Total    string `json:"total"`
	Count    int    `json:"count"`
}

type SpendingSummary struct {
	Total   string `json:"total"`
	Count   int    `json:"count"`
	Average string `json:"average"`
}

// ─── repository ──────────────────────────────────────────────────────────────

type AnalyticsRepository struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

// scopeClause returns the WHERE fragment and args for scoping to user or household.
func scopeClause(f AnalyticsFilters, argOffset int) (string, []any) {
	args := []any{f.From, f.To}
	clause := fmt.Sprintf("date >= $%d AND date <= $%d", argOffset, argOffset+1)
	argOffset += 2

	if f.UserID != nil {
		clause += fmt.Sprintf(" AND scope = 'personal' AND user_id = $%d", argOffset)
		args = append(args, *f.UserID)
	} else {
		clause += fmt.Sprintf(" AND scope = 'household' AND household_id = $%d", argOffset)
		args = append(args, *f.HouseholdID)
	}

	return clause, args
}

// Summary returns total, count, average for the filtered expense set.
func (r *AnalyticsRepository) Summary(ctx context.Context, f AnalyticsFilters) (SpendingSummary, error) {
	where, args := scopeClause(f, 1)
	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(amount), 0)::text,
			COUNT(*)::int,
			COALESCE(AVG(amount), 0)::text
		FROM expenses
		WHERE is_deleted = FALSE AND %s
	`, where)

	var s SpendingSummary
	err := r.pool.QueryRow(ctx, query, args...).Scan(&s.Total, &s.Count, &s.Average)
	if err != nil {
		return SpendingSummary{}, fmt.Errorf("analytics summary: %w", err)
	}
	return s, nil
}

// SpendingByCategory returns per-category totals, sorted by total descending.
func (r *AnalyticsRepository) SpendingByCategory(ctx context.Context, f AnalyticsFilters) ([]SpendingByCategory, error) {
	where, args := scopeClause(f, 1)
	query := fmt.Sprintf(`
		WITH totals AS (
			SELECT
				e.category_id,
				COALESCE(c.name, 'Uncategorised') AS category_name,
				SUM(e.amount)                      AS total,
				COUNT(*)                           AS cnt
			FROM expenses e
			LEFT JOIN categories c ON c.id = e.category_id
			WHERE e.is_deleted = FALSE AND %s
			GROUP BY e.category_id, c.name
		),
		grand AS (SELECT SUM(total) AS grand_total FROM totals)
		SELECT
			t.category_id,
			t.category_name,
			t.total::text,
			t.cnt::int,
			CASE WHEN g.grand_total > 0
				THEN ROUND((t.total / g.grand_total * 100)::numeric, 2)::float
				ELSE 0
			END AS pct
		FROM totals t, grand g
		ORDER BY t.total DESC
	`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("analytics by category: %w", err)
	}
	defer rows.Close()

	var out []SpendingByCategory
	for rows.Next() {
		var row SpendingByCategory
		if err := rows.Scan(&row.CategoryID, &row.CategoryName, &row.Total, &row.Count, &row.PctOfTotal); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// MonthlyTrends returns month-by-month totals for the past N months (default 12).
func (r *AnalyticsRepository) MonthlyTrends(ctx context.Context, f AnalyticsFilters) ([]MonthlyTrend, error) {
	where, args := scopeClause(f, 1)
	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(date_trunc('month', date), 'YYYY-MM') AS month,
			SUM(amount)::text                             AS total,
			COUNT(*)::int                                 AS cnt
		FROM expenses
		WHERE is_deleted = FALSE AND %s
		GROUP BY 1
		ORDER BY 1
	`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("analytics trends: %w", err)
	}
	defer rows.Close()

	var out []MonthlyTrend
	for rows.Next() {
		var row MonthlyTrend
		if err := rows.Scan(&row.Month, &row.Total, &row.Count); err != nil {
			return nil, fmt.Errorf("scan trend row: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// TopMerchants returns the top N merchants by total spend.
func (r *AnalyticsRepository) TopMerchants(ctx context.Context, f AnalyticsFilters, limit int) ([]TopMerchant, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	where, args := scopeClause(f, 1)
	query := fmt.Sprintf(`
		SELECT
			merchant,
			SUM(amount)::text AS total,
			COUNT(*)::int     AS cnt
		FROM expenses
		WHERE is_deleted = FALSE AND %s
		GROUP BY merchant
		ORDER BY SUM(amount) DESC
		LIMIT %d
	`, where, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("analytics merchants: %w", err)
	}
	defer rows.Close()

	var out []TopMerchant
	for rows.Next() {
		var row TopMerchant
		if err := rows.Scan(&row.Merchant, &row.Total, &row.Count); err != nil {
			return nil, fmt.Errorf("scan merchant row: %w", err)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
