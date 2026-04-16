package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var (
	ErrCycleConfigNotFound   = errors.New("no cycle config found for this household")
	ErrCycleSnapshotNotFound = errors.New("cycle snapshot not found")
	ErrInvalidStartDay       = errors.New("start_day must be between 1 and 28")
)

type cycleStore interface {
	CreateConfig(ctx context.Context, params repository.CreateCycleConfigParams) (model.BudgetCycleConfig, error)
	GetActiveConfig(ctx context.Context, householdID string) (model.BudgetCycleConfig, error)
	CreateSnapshot(ctx context.Context, params repository.CreateSnapshotParams) (model.CycleSnapshot, error)
	GetOpenSnapshot(ctx context.Context, householdID string) (model.CycleSnapshot, error)
	GetSnapshotByID(ctx context.Context, snapshotID string) (model.CycleSnapshot, error)
	ListSnapshots(ctx context.Context, householdID string) ([]model.CycleSnapshot, error)
	CloseSnapshot(ctx context.Context, snapshotID string) error
}

// CycleState is the combined view returned to callers.
type CycleState struct {
	Config          model.BudgetCycleConfig `json:"config"`
	CurrentSnapshot model.CycleSnapshot     `json:"current_snapshot"`
}

type BudgetCycleService struct {
	cycles     cycleStore
	households householdStore
	now        func() time.Time
}

func NewBudgetCycleService(cycles cycleStore, households householdStore) *BudgetCycleService {
	return &BudgetCycleService{
		cycles:     cycles,
		households: households,
		now:        time.Now,
	}
}

// GetCycle returns the active config and current open snapshot.
// It lazily creates or rolls over snapshots as needed.
func (s *BudgetCycleService) GetCycle(ctx context.Context, requesterID, householdID string) (CycleState, error) {
	if err := s.requireHouseholdMember(ctx, householdID, requesterID); err != nil {
		return CycleState{}, err
	}

	cfg, err := s.cycles.GetActiveConfig(ctx, householdID)
	if err != nil {
		if errors.Is(err, repository.ErrCycleConfigNotFound) {
			return CycleState{}, ErrCycleConfigNotFound
		}
		return CycleState{}, err
	}

	snapshot, err := s.ensureCurrentSnapshot(ctx, cfg)
	if err != nil {
		return CycleState{}, err
	}

	return CycleState{Config: cfg, CurrentSnapshot: snapshot}, nil
}

// SetCycleConfig creates or updates the cycle config for a household.
// Only owners and editors may change it.
func (s *BudgetCycleService) SetCycleConfig(ctx context.Context, requesterID, householdID string, startDay int) (CycleState, error) {
	if startDay < 1 || startDay > 28 {
		return CycleState{}, ErrInvalidStartDay
	}

	if err := s.requireHouseholdRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return CycleState{}, err
	}

	today := s.now().UTC()
	cfg, err := s.cycles.CreateConfig(ctx, repository.CreateCycleConfigParams{
		HouseholdID:   householdID,
		StartDay:      startDay,
		EffectiveFrom: today.Format("2006-01-02"),
		CreatedBy:     requesterID,
	})
	if err != nil {
		return CycleState{}, fmt.Errorf("set cycle config: %w", err)
	}

	snapshot, err := s.ensureCurrentSnapshot(ctx, cfg)
	if err != nil {
		return CycleState{}, err
	}

	return CycleState{Config: cfg, CurrentSnapshot: snapshot}, nil
}

// ListSnapshots returns all snapshots for a household in descending order.
func (s *BudgetCycleService) ListSnapshots(ctx context.Context, requesterID, householdID string) ([]model.CycleSnapshot, error) {
	if err := s.requireHouseholdMember(ctx, householdID, requesterID); err != nil {
		return nil, err
	}

	return s.cycles.ListSnapshots(ctx, householdID)
}

// ── Cycle math ────────────────────────────────────────────────────────────────

// ensureCurrentSnapshot guarantees there is a current, open snapshot.
// If the existing open snapshot has expired it is closed and a new one is created.
func (s *BudgetCycleService) ensureCurrentSnapshot(ctx context.Context, cfg model.BudgetCycleConfig) (model.CycleSnapshot, error) {
	today := s.now().UTC()
	wantStart, wantEnd := currentCycleBounds(cfg.StartDay, today)

	existing, err := s.cycles.GetOpenSnapshot(ctx, cfg.HouseholdID)
	if err != nil && !errors.Is(err, repository.ErrCycleSnapshotNotFound) {
		return model.CycleSnapshot{}, err
	}

	// If there's an open snapshot that matches the current period, return it.
	if err == nil && existing.CycleStart == wantStart.Format("2006-01-02") {
		return existing, nil
	}

	// Close stale open snapshot if present.
	if err == nil {
		if closeErr := s.cycles.CloseSnapshot(ctx, existing.ID); closeErr != nil {
			return model.CycleSnapshot{}, closeErr
		}
	}

	// Create the snapshot for the current period.
	snapshot, createErr := s.cycles.CreateSnapshot(ctx, repository.CreateSnapshotParams{
		HouseholdID: cfg.HouseholdID,
		CycleStart:  wantStart.Format("2006-01-02"),
		CycleEnd:    wantEnd.Format("2006-01-02"),
		Label:       cycleLabel(wantStart, wantEnd),
		ConfigID:    cfg.ID,
	})
	if createErr != nil {
		return model.CycleSnapshot{}, createErr
	}

	return snapshot, nil
}

// currentCycleBounds returns the start and end dates of the billing cycle that
// contains ref, given a monthly start_day.
//
// Example: startDay=25, ref=April 13 → March 25 – April 24
//
//	startDay=25, ref=April 27 → April 25 – May 24
func currentCycleBounds(startDay int, ref time.Time) (start, end time.Time) {
	year, month, day := ref.Date()
	loc := ref.Location()

	if day >= startDay {
		start = time.Date(year, month, startDay, 0, 0, 0, 0, loc)
	} else {
		prevMonth := month - 1
		prevYear := year
		if prevMonth == 0 {
			prevMonth = 12
			prevYear--
		}
		start = time.Date(prevYear, prevMonth, startDay, 0, 0, 0, 0, loc)
	}

	// end = start + 1 month - 1 day
	end = start.AddDate(0, 1, -1)
	return start, end
}

// cycleLabel produces a human-readable label for a cycle period.
// Same month+year: "April 2026"
// Spans two months: "Mar 25 – Apr 24, 2026"
func cycleLabel(start, end time.Time) string {
	if start.Month() == end.Month() && start.Year() == end.Year() {
		return start.Format("January 2006")
	}

	// Use abbreviated month names with day numbers.
	startPart := fmt.Sprintf("%s %d", start.Format("Jan"), start.Day())
	endYear := end.Year()
	endPart := fmt.Sprintf("%s %d, %d", end.Format("Jan"), end.Day(), endYear)
	return startPart + " – " + endPart
}

// ── Auth helpers ──────────────────────────────────────────────────────────────

func (s *BudgetCycleService) requireHouseholdMember(ctx context.Context, householdID, userID string) error {
	_, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotMember
		}
		return err
	}
	return nil
}

func (s *BudgetCycleService) requireHouseholdRole(ctx context.Context, householdID, userID string, roles ...string) error {
	m, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotMember
		}
		return err
	}
	for _, r := range roles {
		if m.Role == r {
			return nil
		}
	}
	return ErrInsufficientRole
}
