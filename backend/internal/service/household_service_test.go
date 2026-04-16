package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

func TestHouseholdServiceAcceptInvitationRequiresMatchingEmail(t *testing.T) {
	t.Parallel()

	service := NewHouseholdService(&invitationHouseholdStoreStub{
		invitation: model.Invitation{
			ID:          "inv-1",
			HouseholdID: "household-1",
			Email:       "invited@example.com",
			Role:        "viewer",
			ExpiresAt:   time.Now().Add(time.Hour),
		},
	}, &invitationUserStoreStub{
		user: model.User{ID: "user-1", Email: "other@example.com"},
	}, auth.NewTokenService("12345678901234567890123456789012"), time.Hour)

	_, err := service.AcceptInvitation(context.Background(), "user-1", "plain-token")
	if !errors.Is(err, ErrInvitationEmailMismatch) {
		t.Fatalf("expected ErrInvitationEmailMismatch, got %v", err)
	}
}

type invitationHouseholdStoreStub struct {
	invitation model.Invitation
}

func (s *invitationHouseholdStoreStub) Create(context.Context, string, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *invitationHouseholdStoreStub) ListByUserID(context.Context, string) ([]model.Household, error) {
	return nil, nil
}

func (s *invitationHouseholdStoreStub) GetByID(context.Context, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *invitationHouseholdStoreStub) Update(context.Context, string, string) (model.Household, error) {
	return model.Household{}, nil
}

func (s *invitationHouseholdStoreStub) Delete(context.Context, string) error {
	return nil
}

func (s *invitationHouseholdStoreStub) GetMembership(context.Context, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}

func (s *invitationHouseholdStoreStub) ListMembers(context.Context, string) ([]model.HouseholdMember, error) {
	return nil, nil
}

func (s *invitationHouseholdStoreStub) UpdateMemberRole(context.Context, string, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}

func (s *invitationHouseholdStoreStub) RemoveMember(context.Context, string, string) error {
	return nil
}

func (s *invitationHouseholdStoreStub) CountOwners(context.Context, string) (int, error) {
	return 0, nil
}

func (s *invitationHouseholdStoreStub) CreateInvitation(context.Context, repository.CreateInvitationParams) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (s *invitationHouseholdStoreStub) ListPendingInvitations(context.Context, string) ([]model.Invitation, error) {
	return nil, nil
}

func (s *invitationHouseholdStoreStub) UpdateInvitationStatus(context.Context, string, string, string) error {
	return nil
}

func (s *invitationHouseholdStoreStub) FindPendingInvitationByTokenHash(context.Context, string) (model.Invitation, error) {
	return s.invitation, nil
}

func (s *invitationHouseholdStoreStub) AcceptInvitation(context.Context, string, string, string, string) (model.HouseholdMember, error) {
	return model.HouseholdMember{}, nil
}

type invitationUserStoreStub struct {
	user model.User
	err  error
}

func (s *invitationUserStoreStub) FindByID(context.Context, string) (model.User, error) {
	if s.err != nil {
		return model.User{}, s.err
	}
	return s.user, nil
}
