package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

func TestCategoryServiceListRequiresMembership(t *testing.T) {
	t.Parallel()

	service := NewCategoryService(&categoryStoreStub{}, &categoryHouseholdStoreStub{
		membershipErr: repository.ErrMemberNotFound,
	})
	householdID := "household-1"

	_, err := service.List(context.Background(), "user-1", &householdID)
	if !errors.Is(err, ErrNotMember) {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestCategoryServiceCreateRequiresEditorOrOwner(t *testing.T) {
	t.Parallel()

	service := NewCategoryService(&categoryStoreStub{}, &categoryHouseholdStoreStub{
		member: model.HouseholdMember{Role: "viewer"},
	})

	_, err := service.Create(context.Background(), "user-1", "household-1", "Food", nil, nil, nil)
	if !errors.Is(err, ErrInsufficientRole) {
		t.Fatalf("expected ErrInsufficientRole, got %v", err)
	}
}

type categoryStoreStub struct{}

func (s *categoryStoreStub) List(context.Context, *string) ([]model.Category, error) {
	return nil, nil
}

func (s *categoryStoreStub) GetByID(context.Context, string) (model.Category, error) {
	return model.Category{}, nil
}

func (s *categoryStoreStub) Create(context.Context, repository.CreateCategoryParams) (model.Category, error) {
	return model.Category{}, nil
}

func (s *categoryStoreStub) Update(context.Context, repository.UpdateCategoryParams) (model.Category, error) {
	return model.Category{}, nil
}

func (s *categoryStoreStub) Delete(context.Context, string, string) error {
	return nil
}

type categoryHouseholdStoreStub struct {
	member        model.HouseholdMember
	membershipErr error
}

func (s *categoryHouseholdStoreStub) Create(ctx context.Context, name, createdByUserID string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *categoryHouseholdStoreStub) ListByUserID(context.Context, string) ([]model.Household, error) {
	return nil, nil
}

func (s *categoryHouseholdStoreStub) GetByID(context.Context, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *categoryHouseholdStoreStub) Update(context.Context, string, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *categoryHouseholdStoreStub) Delete(context.Context, string) error {
	return nil
}

func (s *categoryHouseholdStoreStub) GetMembership(context.Context, string, string) (model.HouseholdMember, error) {
	if s.membershipErr != nil {
		return model.HouseholdMember{}, s.membershipErr
	}
	return s.member, nil
}

func (s *categoryHouseholdStoreStub) ListMembers(context.Context, string) ([]model.HouseholdMember, error) {
	return nil, nil
}

func (s *categoryHouseholdStoreStub) UpdateMemberRole(context.Context, string, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}

func (s *categoryHouseholdStoreStub) RemoveMember(context.Context, string, string) error {
	return nil
}

func (s *categoryHouseholdStoreStub) CountOwners(context.Context, string) (int, error) {
	return 0, nil
}

func (s *categoryHouseholdStoreStub) CreateInvitation(context.Context, repository.CreateInvitationParams) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (s *categoryHouseholdStoreStub) ListPendingInvitations(context.Context, string) ([]model.Invitation, error) {
	return nil, nil
}

func (s *categoryHouseholdStoreStub) UpdateInvitationStatus(context.Context, string, string, string) error {
	return nil
}

func (s *categoryHouseholdStoreStub) FindPendingInvitationByTokenHash(context.Context, string) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (s *categoryHouseholdStoreStub) AcceptInvitation(context.Context, string, string, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}
