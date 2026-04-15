package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

func TestBudgetServiceCreateHouseholdBudgetRequiresEditorOrOwner(t *testing.T) {
	t.Parallel()

	service := NewBudgetService(&budgetStoreStub{}, &snapshotGetterStub{}, &budgetHouseholdStoreStub{
		member: model.HouseholdMember{Role: "viewer"},
	}, &thresholdCheckerStub{})

	_, err := service.Create(context.Background(), "user-1", CreateBudgetInput{
		HouseholdID: "household-1",
		Scope:       "household",
		Amount:      "100",
	})
	if !errors.Is(err, ErrInsufficientRole) {
		t.Fatalf("expected ErrInsufficientRole, got %v", err)
	}
}

func TestBudgetServiceRejectsSnapshotFromAnotherHousehold(t *testing.T) {
	t.Parallel()

	snapshotID := "snapshot-1"
	service := NewBudgetService(&budgetStoreStub{}, &snapshotGetterStub{
		snapshot: model.CycleSnapshot{ID: snapshotID, HouseholdID: "household-2"},
	}, &budgetHouseholdStoreStub{
		member: model.HouseholdMember{Role: "owner"},
	}, &thresholdCheckerStub{})

	_, err := service.Create(context.Background(), "user-1", CreateBudgetInput{
		HouseholdID: "household-1",
		Scope:       "household",
		SnapshotID:  &snapshotID,
		Amount:      "100",
	})
	if !errors.Is(err, ErrCycleSnapshotNotFound) {
		t.Fatalf("expected ErrCycleSnapshotNotFound, got %v", err)
	}
}

type budgetStoreStub struct{}

func (s *budgetStoreStub) Create(context.Context, repository.CreateBudgetParams) (model.Budget, error) {
	return model.Budget{}, nil
}

func (s *budgetStoreStub) List(context.Context, repository.ListBudgetFilters) ([]model.Budget, error) {
	return nil, nil
}

func (s *budgetStoreStub) GetByID(context.Context, string) (model.Budget, error) {
	return model.Budget{}, nil
}

func (s *budgetStoreStub) Update(context.Context, repository.UpdateBudgetParams) (model.Budget, error) {
	return model.Budget{}, nil
}

func (s *budgetStoreStub) Delete(context.Context, string) error {
	return nil
}

func (s *budgetStoreStub) GetHealthData(context.Context, string, string, *string, *string) ([]repository.BudgetHealthRow, error) {
	return nil, nil
}

type snapshotGetterStub struct {
	snapshot model.CycleSnapshot
	err      error
}

func (s *snapshotGetterStub) GetOpenSnapshot(context.Context, string) (model.CycleSnapshot, error) {
	return model.CycleSnapshot{ID: "open-snapshot", HouseholdID: "household-1"}, nil
}

func (s *snapshotGetterStub) GetSnapshotByID(context.Context, string) (model.CycleSnapshot, error) {
	if s.err != nil {
		return model.CycleSnapshot{}, s.err
	}
	return s.snapshot, nil
}

type budgetHouseholdStoreStub struct {
	member        model.HouseholdMember
	membershipErr error
}

func (s *budgetHouseholdStoreStub) Create(context.Context, string, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *budgetHouseholdStoreStub) ListByUserID(context.Context, string) ([]model.Household, error) {
	return nil, nil
}

func (s *budgetHouseholdStoreStub) GetByID(context.Context, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *budgetHouseholdStoreStub) Update(context.Context, string, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *budgetHouseholdStoreStub) Delete(context.Context, string) error {
	return nil
}

func (s *budgetHouseholdStoreStub) GetMembership(context.Context, string, string) (model.HouseholdMember, error) {
	if s.membershipErr != nil {
		return model.HouseholdMember{}, s.membershipErr
	}
	return s.member, nil
}

func (s *budgetHouseholdStoreStub) ListMembers(context.Context, string) ([]model.HouseholdMember, error) {
	return nil, nil
}

func (s *budgetHouseholdStoreStub) UpdateMemberRole(context.Context, string, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}

func (s *budgetHouseholdStoreStub) RemoveMember(context.Context, string, string) error {
	return nil
}

func (s *budgetHouseholdStoreStub) CountOwners(context.Context, string) (int, error) {
	return 0, nil
}

func (s *budgetHouseholdStoreStub) CreateInvitation(context.Context, repository.CreateInvitationParams) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (s *budgetHouseholdStoreStub) ListPendingInvitations(context.Context, string) ([]model.Invitation, error) {
	return nil, nil
}

func (s *budgetHouseholdStoreStub) UpdateInvitationStatus(context.Context, string, string, string) error {
	return nil
}

func (s *budgetHouseholdStoreStub) FindPendingInvitationByTokenHash(context.Context, string) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (s *budgetHouseholdStoreStub) AcceptInvitation(context.Context, string, string, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}

type thresholdCheckerStub struct{}

func (s *thresholdCheckerStub) CheckBudgetThresholds(context.Context, string, []model.BudgetHealthItem, string) {}
