package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var (
	ErrHouseholdNotFound     = errors.New("household not found")
	ErrMemberNotFound        = errors.New("household member not found")
	ErrAlreadyMember         = errors.New("user is already a member of this household")
	ErrNotMember             = errors.New("user is not a member of this household")
	ErrInsufficientRole      = errors.New("insufficient role for this action")
	ErrCannotRemoveLastOwner = errors.New("cannot remove the last owner of a household")
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationExpired     = errors.New("invitation has expired")
	ErrInvitationConflict    = errors.New("a pending invitation for this email already exists")
	ErrInvitationEmailMismatch = errors.New("invitation email does not match the authenticated user")
)

type householdStore interface {
	Create(ctx context.Context, name, createdByUserID string) (model.Household, error)
	ListByUserID(ctx context.Context, userID string) ([]model.Household, error)
	GetByID(ctx context.Context, householdID string) (model.Household, error)
	Update(ctx context.Context, householdID, name string) (model.Household, error)
	Delete(ctx context.Context, householdID string) error

	GetMembership(ctx context.Context, householdID, userID string) (model.HouseholdMember, error)
	ListMembers(ctx context.Context, householdID string) ([]model.HouseholdMember, error)
	UpdateMemberRole(ctx context.Context, householdID, userID, role string) (model.HouseholdMember, error)
	RemoveMember(ctx context.Context, householdID, userID string) error
	CountOwners(ctx context.Context, householdID string) (int, error)

	CreateInvitation(ctx context.Context, params repository.CreateInvitationParams) (model.Invitation, error)
	ListPendingInvitations(ctx context.Context, householdID string) ([]model.Invitation, error)
	UpdateInvitationStatus(ctx context.Context, invitationID, householdID, status string) error
	FindPendingInvitationByTokenHash(ctx context.Context, tokenHash string) (model.Invitation, error)
	AcceptInvitation(ctx context.Context, invitationID, householdID, userID, role string) (model.HouseholdMember, error)
}

type householdUserStore interface {
	FindByID(ctx context.Context, id string) (model.User, error)
}

type HouseholdService struct {
	households householdStore
	users      householdUserStore
	tokens     *auth.TokenService
	inviteTTL  time.Duration
	now        func() time.Time
}

func NewHouseholdService(households householdStore, users householdUserStore, tokens *auth.TokenService, inviteTTL time.Duration) *HouseholdService {
	return &HouseholdService{
		households: households,
		users:      users,
		tokens:     tokens,
		inviteTTL:  inviteTTL,
		now:        time.Now,
	}
}

// ── Household CRUD ────────────────────────────────────────────────────────────

func (s *HouseholdService) Create(ctx context.Context, requesterID, name string) (model.Household, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Household{}, errors.New("name is required")
	}

	return s.households.Create(ctx, name, requesterID)
}

func (s *HouseholdService) List(ctx context.Context, userID string) ([]model.Household, error) {
	return s.households.ListByUserID(ctx, userID)
}

func (s *HouseholdService) Get(ctx context.Context, requesterID, householdID string) (model.Household, error) {
	if _, err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return model.Household{}, err
	}

	h, err := s.households.GetByID(ctx, householdID)
	if err != nil {
		if errors.Is(err, repository.ErrHouseholdNotFound) {
			return model.Household{}, ErrHouseholdNotFound
		}
		return model.Household{}, err
	}

	return h, nil
}

func (s *HouseholdService) Update(ctx context.Context, requesterID, householdID, name string) (model.Household, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Household{}, errors.New("name is required")
	}

	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return model.Household{}, err
	}

	h, err := s.households.Update(ctx, householdID, name)
	if err != nil {
		if errors.Is(err, repository.ErrHouseholdNotFound) {
			return model.Household{}, ErrHouseholdNotFound
		}
		return model.Household{}, err
	}

	return h, nil
}

func (s *HouseholdService) Delete(ctx context.Context, requesterID, householdID string) error {
	if err := s.requireRole(ctx, householdID, requesterID, "owner"); err != nil {
		return err
	}

	if err := s.households.Delete(ctx, householdID); err != nil {
		if errors.Is(err, repository.ErrHouseholdNotFound) {
			return ErrHouseholdNotFound
		}
		return err
	}

	return nil
}

func (s *HouseholdService) Leave(ctx context.Context, requesterID, householdID string) error {
	if _, err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return err
	}

	if err := s.guardLastOwner(ctx, householdID, requesterID); err != nil {
		return err
	}

	return s.households.RemoveMember(ctx, householdID, requesterID)
}

// ── Members ───────────────────────────────────────────────────────────────────

func (s *HouseholdService) ListMembers(ctx context.Context, requesterID, householdID string) ([]model.HouseholdMember, error) {
	if _, err := s.requireMember(ctx, householdID, requesterID); err != nil {
		return nil, err
	}

	return s.households.ListMembers(ctx, householdID)
}

func (s *HouseholdService) UpdateMemberRole(ctx context.Context, requesterID, householdID, targetUserID, role string) (model.HouseholdMember, error) {
	if err := s.requireRole(ctx, householdID, requesterID, "owner"); err != nil {
		return model.HouseholdMember{}, err
	}

	if requesterID == targetUserID {
		return model.HouseholdMember{}, errors.New("cannot change your own role")
	}

	if !isValidRole(role) {
		return model.HouseholdMember{}, fmt.Errorf("invalid role: must be owner, editor, or viewer")
	}

	m, err := s.households.UpdateMemberRole(ctx, householdID, targetUserID, role)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return model.HouseholdMember{}, ErrMemberNotFound
		}
		return model.HouseholdMember{}, err
	}

	return m, nil
}

func (s *HouseholdService) RemoveMember(ctx context.Context, requesterID, householdID, targetUserID string) error {
	if err := s.requireRole(ctx, householdID, requesterID, "owner"); err != nil {
		return err
	}

	if requesterID == targetUserID {
		return errors.New("use the leave endpoint to remove yourself")
	}

	if err := s.guardLastOwner(ctx, householdID, targetUserID); err != nil {
		return err
	}

	if err := s.households.RemoveMember(ctx, householdID, targetUserID); err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrMemberNotFound
		}
		return err
	}

	return nil
}

// ── Invitations ───────────────────────────────────────────────────────────────

type InviteResult struct {
	Invitation model.Invitation `json:"invitation"`
	Token      string           `json:"token"` // plain token for the inviter to forward
}

func (s *HouseholdService) CreateInvitation(ctx context.Context, requesterID, householdID, email, role string) (InviteResult, error) {
	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return InviteResult{}, err
	}

	if !isValidRole(role) {
		return InviteResult{}, fmt.Errorf("invalid role: must be owner, editor, or viewer")
	}

	plainToken, tokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return InviteResult{}, fmt.Errorf("generate invite token: %w", err)
	}

	inv, err := s.households.CreateInvitation(ctx, repository.CreateInvitationParams{
		HouseholdID: householdID,
		Email:       strings.TrimSpace(strings.ToLower(email)),
		Role:        role,
		TokenHash:   tokenHash,
		ExpiresAt:   s.now().Add(s.inviteTTL),
	})
	if err != nil {
		if errors.Is(err, repository.ErrInvitationConflict) {
			return InviteResult{}, ErrInvitationConflict
		}
		return InviteResult{}, err
	}

	return InviteResult{Invitation: inv, Token: plainToken}, nil
}

func (s *HouseholdService) ListInvitations(ctx context.Context, requesterID, householdID string) ([]model.Invitation, error) {
	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return nil, err
	}

	return s.households.ListPendingInvitations(ctx, householdID)
}

func (s *HouseholdService) RevokeInvitation(ctx context.Context, requesterID, householdID, invitationID string) error {
	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return err
	}

	if err := s.households.UpdateInvitationStatus(ctx, invitationID, householdID, "revoked"); err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	return nil
}

func (s *HouseholdService) AcceptInvitation(ctx context.Context, userID, token string) (model.HouseholdMember, error) {
	tokenHash := auth.HashToken(strings.TrimSpace(token))

	inv, err := s.households.FindPendingInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return model.HouseholdMember{}, ErrInvitationNotFound
		}
		return model.HouseholdMember{}, err
	}

	if !inv.ExpiresAt.After(s.now()) {
		return model.HouseholdMember{}, ErrInvitationExpired
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.HouseholdMember{}, ErrInvitationNotFound
		}
		return model.HouseholdMember{}, err
	}

	if !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(inv.Email)) {
		return model.HouseholdMember{}, ErrInvitationEmailMismatch
	}

	m, err := s.households.AcceptInvitation(ctx, inv.ID, inv.HouseholdID, userID, inv.Role)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyMember) {
			return model.HouseholdMember{}, ErrAlreadyMember
		}
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return model.HouseholdMember{}, ErrInvitationNotFound
		}
		return model.HouseholdMember{}, err
	}

	return m, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *HouseholdService) requireMember(ctx context.Context, householdID, userID string) (model.HouseholdMember, error) {
	m, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return model.HouseholdMember{}, ErrNotMember
		}
		return model.HouseholdMember{}, err
	}

	return m, nil
}

func (s *HouseholdService) requireRole(ctx context.Context, householdID, userID string, roles ...string) error {
	m, err := s.requireMember(ctx, householdID, userID)
	if err != nil {
		return err
	}

	for _, r := range roles {
		if m.Role == r {
			return nil
		}
	}

	return ErrInsufficientRole
}

func (s *HouseholdService) guardLastOwner(ctx context.Context, householdID, userID string) error {
	m, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		// Not a member — no concern.
		return nil
	}

	if m.Role != "owner" {
		return nil
	}

	count, err := s.households.CountOwners(ctx, householdID)
	if err != nil {
		return err
	}

	if count <= 1 {
		return ErrCannotRemoveLastOwner
	}

	return nil
}

func isValidRole(role string) bool {
	return role == "owner" || role == "editor" || role == "viewer"
}
