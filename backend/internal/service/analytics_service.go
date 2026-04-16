package service

import (
	"context"
	"errors"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

type analyticsStore interface {
	Summary(ctx context.Context, f repository.AnalyticsFilters) (repository.SpendingSummary, error)
	SpendingByCategory(ctx context.Context, f repository.AnalyticsFilters) ([]repository.SpendingByCategory, error)
	MonthlyTrends(ctx context.Context, f repository.AnalyticsFilters) ([]repository.MonthlyTrend, error)
	TopMerchants(ctx context.Context, f repository.AnalyticsFilters, limit int) ([]repository.TopMerchant, error)
}

type AnalyticsService struct {
	store      analyticsStore
	households householdStore
}

func NewAnalyticsService(store analyticsStore, households householdStore) *AnalyticsService {
	return &AnalyticsService{store: store, households: households}
}

type AnalyticsInput struct {
	// Scope is "personal" or "household".
	Scope       string
	UserID      string
	HouseholdID string // required when Scope == "household"
	From        string // YYYY-MM-DD; defaults to start of current month
	To          string // YYYY-MM-DD; defaults to end of current month
	Limit       int    // only used by merchants
}

func (s *AnalyticsService) buildFilters(ctx context.Context, input AnalyticsInput) (repository.AnalyticsFilters, error) {
	from, to, err := parseDateRange(input.From, input.To)
	if err != nil {
		return repository.AnalyticsFilters{}, err
	}

	f := repository.AnalyticsFilters{From: from, To: to}

	if input.Scope == "household" {
		if input.HouseholdID == "" {
			return repository.AnalyticsFilters{}, errors.New("household_id required for household scope")
		}
		if _, err := s.households.GetMembership(ctx, input.HouseholdID, input.UserID); err != nil {
			if errors.Is(err, repository.ErrMemberNotFound) {
				return repository.AnalyticsFilters{}, ErrNotMember
			}
			return repository.AnalyticsFilters{}, err
		}
		f.HouseholdID = &input.HouseholdID
	} else {
		f.UserID = &input.UserID
	}

	return f, nil
}

func (s *AnalyticsService) GetSpending(ctx context.Context, input AnalyticsInput) (
	summary repository.SpendingSummary,
	byCategory []repository.SpendingByCategory,
	err error,
) {
	f, err := s.buildFilters(ctx, input)
	if err != nil {
		return repository.SpendingSummary{}, nil, err
	}

	summary, err = s.store.Summary(ctx, f)
	if err != nil {
		return repository.SpendingSummary{}, nil, err
	}

	byCategory, err = s.store.SpendingByCategory(ctx, f)
	if err != nil {
		return repository.SpendingSummary{}, nil, err
	}
	if byCategory == nil {
		byCategory = []repository.SpendingByCategory{}
	}

	return summary, byCategory, nil
}

func (s *AnalyticsService) GetTrends(ctx context.Context, input AnalyticsInput) ([]repository.MonthlyTrend, error) {
	f, err := s.buildFilters(ctx, input)
	if err != nil {
		return nil, err
	}

	trends, err := s.store.MonthlyTrends(ctx, f)
	if err != nil {
		return nil, err
	}
	if trends == nil {
		trends = []repository.MonthlyTrend{}
	}
	return trends, nil
}

func (s *AnalyticsService) GetTopMerchants(ctx context.Context, input AnalyticsInput) ([]repository.TopMerchant, error) {
	f, err := s.buildFilters(ctx, input)
	if err != nil {
		return nil, err
	}

	merchants, err := s.store.TopMerchants(ctx, f, input.Limit)
	if err != nil {
		return nil, err
	}
	if merchants == nil {
		merchants = []repository.TopMerchant{}
	}
	return merchants, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func parseDateRange(from, to string) (start, end time.Time, err error) {
	now := time.Now()

	var f, t time.Time

	if from == "" {
		f = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		f, err = time.Parse("2006-01-02", from)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("from must be YYYY-MM-DD")
		}
	}

	if to == "" {
		// last day of current month
		t = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.UTC)
	} else {
		t, err = time.Parse("2006-01-02", to)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("to must be YYYY-MM-DD")
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	if f.After(t) {
		return time.Time{}, time.Time{}, errors.New("from must be before or equal to to")
	}

	return f, t, nil
}
