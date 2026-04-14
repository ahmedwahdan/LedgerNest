package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var (
	ErrBudgetNotFound  = errors.New("budget not found")
	ErrBudgetConflict  = errors.New("a budget for this category already exists in this snapshot")
	ErrBudgetForbidden = errors.New("you don't have access to this budget")
)

type budgetStore interface {
	Create(ctx context.Context, params repository.CreateBudgetParams) (model.Budget, error)
	List(ctx context.Context, filters repository.ListBudgetFilters) ([]model.Budget, error)
	GetByID(ctx context.Context, budgetID string) (model.Budget, error)
	Update(ctx context.Context, params repository.UpdateBudgetParams) (model.Budget, error)
	Delete(ctx context.Context, budgetID string) error
	GetHealthData(ctx context.Context, snapshotID, scope string, userID, householdID *string) ([]repository.BudgetHealthRow, error)
}

type snapshotGetter interface {
	GetOpenSnapshot(ctx context.Context, householdID string) (model.CycleSnapshot, error)
	GetSnapshotByID(ctx context.Context, snapshotID string) (model.CycleSnapshot, error)
}

type thresholdChecker interface {
	CheckBudgetThresholds(ctx context.Context, userID string, items []model.BudgetHealthItem, snapshotID string)
}

type BudgetService struct {
	budgets       budgetStore
	cycles        snapshotGetter
	households    householdStore
	notifications thresholdChecker
}

func NewBudgetService(budgets budgetStore, cycles snapshotGetter, households householdStore, notifications thresholdChecker) *BudgetService {
	return &BudgetService{
		budgets:       budgets,
		cycles:        cycles,
		households:    households,
		notifications: notifications,
	}
}

// List returns budgets for a snapshot. snapshot_id defaults to the current open one.
func (s *BudgetService) List(ctx context.Context, requesterID, householdID string, snapshotID *string, scope *string) ([]model.Budget, error) {
	if err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return nil, err
	}

	sid, err := s.resolveSnapshotID(ctx, householdID, snapshotID)
	if err != nil {
		return nil, err
	}

	userID := requesterID
	filters := repository.ListBudgetFilters{
		CycleSnapshotID: sid,
		HouseholdID:     &householdID,
	}
	if scope != nil {
		filters.Scope = scope
		if *scope == "personal" {
			filters.UserID = &userID
			filters.HouseholdID = nil
		}
	}

	return s.budgets.List(ctx, filters)
}

type CreateBudgetInput struct {
	HouseholdID string
	Scope       string
	CategoryID  *string
	SnapshotID  *string // nil = current open
	Amount      string
}

func (s *BudgetService) Create(ctx context.Context, requesterID string, input CreateBudgetInput) (model.Budget, error) {
	if err := validateAmount(input.Amount); err != nil {
		return model.Budget{}, err
	}

	if input.Scope != "personal" && input.Scope != "household" {
		return model.Budget{}, errors.New("scope must be personal or household")
	}

	if err := s.requireMember(ctx, input.HouseholdID, requesterID); err != nil {
		return model.Budget{}, err
	}

	sid, err := s.resolveSnapshotID(ctx, input.HouseholdID, input.SnapshotID)
	if err != nil {
		return model.Budget{}, err
	}

	params := repository.CreateBudgetParams{
		Scope:           input.Scope,
		CategoryID:      input.CategoryID,
		CycleSnapshotID: sid,
		Amount:          input.Amount,
	}

	if input.Scope == "personal" {
		params.UserID = &requesterID
	} else {
		params.HouseholdID = &input.HouseholdID
	}

	b, err := s.budgets.Create(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrBudgetConflict) {
			return model.Budget{}, ErrBudgetConflict
		}
		return model.Budget{}, err
	}

	return b, nil
}

func (s *BudgetService) Update(ctx context.Context, requesterID, householdID, budgetID, amount string) (model.Budget, error) {
	if err := validateAmount(amount); err != nil {
		return model.Budget{}, err
	}

	if err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return model.Budget{}, err
	}

	// Verify the budget belongs to this household/user.
	existing, err := s.budgets.GetByID(ctx, budgetID)
	if err != nil {
		if errors.Is(err, repository.ErrBudgetNotFound) {
			return model.Budget{}, ErrBudgetNotFound
		}
		return model.Budget{}, err
	}

	if err := s.checkBudgetAccess(existing, requesterID, householdID); err != nil {
		return model.Budget{}, err
	}

	b, err := s.budgets.Update(ctx, repository.UpdateBudgetParams{
		BudgetID: budgetID,
		Amount:   amount,
	})
	if err != nil {
		if errors.Is(err, repository.ErrBudgetNotFound) {
			return model.Budget{}, ErrBudgetNotFound
		}
		return model.Budget{}, err
	}

	return b, nil
}

func (s *BudgetService) Delete(ctx context.Context, requesterID, householdID, budgetID string) error {
	if err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return err
	}

	existing, err := s.budgets.GetByID(ctx, budgetID)
	if err != nil {
		if errors.Is(err, repository.ErrBudgetNotFound) {
			return ErrBudgetNotFound
		}
		return err
	}

	if err := s.checkBudgetAccess(existing, requesterID, householdID); err != nil {
		return err
	}

	return s.budgets.Delete(ctx, budgetID)
}

// GetHealth returns the health summary for the current (or specified) snapshot.
func (s *BudgetService) GetHealth(ctx context.Context, requesterID, householdID string, snapshotID *string, scope string) (model.BudgetHealth, error) {
	if err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return model.BudgetHealth{}, err
	}

	sid, err := s.resolveSnapshotID(ctx, householdID, snapshotID)
	if err != nil {
		return model.BudgetHealth{}, err
	}

	snapshot, err := s.cycles.GetSnapshotByID(ctx, sid)
	if err != nil {
		if errors.Is(err, repository.ErrCycleSnapshotNotFound) {
			return model.BudgetHealth{}, ErrCycleSnapshotNotFound
		}
		return model.BudgetHealth{}, err
	}

	var userID, hhID *string
	if scope == "personal" {
		userID = &requesterID
	} else {
		hhID = &householdID
		scope = "household"
	}

	rows, err := s.budgets.GetHealthData(ctx, sid, scope, userID, hhID)
	if err != nil {
		return model.BudgetHealth{}, err
	}

	health := model.BudgetHealth{
		Snapshot:   snapshot,
		Categories: make([]model.BudgetHealthItem, 0),
	}

	var allItems []model.BudgetHealthItem
	for _, row := range rows {
		item := buildHealthItem(row)
		if row.CategoryID == nil {
			health.Overall = &item
		} else {
			health.Categories = append(health.Categories, item)
		}
		allItems = append(allItems, item)
	}

	// Fire threshold notifications best-effort (personal scope only — household
	// notifications require knowing which user to notify).
	if scope == "personal" {
		go s.notifications.CheckBudgetThresholds(ctx, requesterID, allItems, sid)
	}

	return health, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *BudgetService) resolveSnapshotID(ctx context.Context, householdID string, snapshotID *string) (string, error) {
	if snapshotID != nil && *snapshotID != "" {
		return *snapshotID, nil
	}

	snap, err := s.cycles.GetOpenSnapshot(ctx, householdID)
	if err != nil {
		if errors.Is(err, repository.ErrCycleSnapshotNotFound) {
			return "", fmt.Errorf("no open cycle snapshot — set a cycle config first")
		}
		return "", err
	}

	return snap.ID, nil
}

func (s *BudgetService) requireMember(ctx context.Context, householdID, userID string) error {
	_, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotMember
		}
		return err
	}
	return nil
}

func (s *BudgetService) checkBudgetAccess(b model.Budget, requesterID, householdID string) error {
	if b.Scope == "personal" {
		if b.UserID == nil || *b.UserID != requesterID {
			return ErrBudgetForbidden
		}
	} else {
		if b.HouseholdID == nil || *b.HouseholdID != householdID {
			return ErrBudgetForbidden
		}
	}
	return nil
}

func buildHealthItem(row repository.BudgetHealthRow) model.BudgetHealthItem {
	amount, _ := strconv.ParseFloat(row.Amount, 64)
	rollover, _ := strconv.ParseFloat(row.Rollover, 64)
	spent, _ := strconv.ParseFloat(row.Spent, 64)

	effective := amount + rollover
	remaining := effective - spent
	pct := 0.0
	if effective > 0 {
		pct = (spent / effective) * 100
	}

	return model.BudgetHealthItem{
		BudgetID:     row.BudgetID,
		CategoryID:   row.CategoryID,
		CategoryName: row.CategoryName,
		Amount:       row.Amount,
		Rollover:     row.Rollover,
		Spent:        fmt.Sprintf("%.2f", spent),
		Remaining:    fmt.Sprintf("%.2f", remaining),
		PctUsed:      pct,
	}
}

func validateAmount(amount string) error {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return errors.New("amount is required")
	}
	v, err := strconv.ParseFloat(amount, 64)
	if err != nil || v < 0 {
		return errors.New("amount must be a non-negative number")
	}
	return nil
}
