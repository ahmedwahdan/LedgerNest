package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

var (
	ErrHouseholdNotFound     = errors.New("household not found")
	ErrMemberNotFound        = errors.New("household member not found")
	ErrAlreadyMember         = errors.New("user is already a member of this household")
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationConflict    = errors.New("a pending invitation for this email already exists")
)

type HouseholdRepository struct {
	pool *pgxpool.Pool
}

func NewHouseholdRepository(pool *pgxpool.Pool) *HouseholdRepository {
	return &HouseholdRepository{pool: pool}
}

// ── Households ────────────────────────────────────────────────────────────────

func (r *HouseholdRepository) Create(ctx context.Context, name, createdByUserID string) (model.Household, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Household{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const createHousehold = `
		INSERT INTO households (name, created_by) VALUES ($1, $2)
		RETURNING id, name, created_by, created_at, updated_at
	`

	var h model.Household
	if err := tx.QueryRow(ctx, createHousehold, name, createdByUserID).Scan(
		&h.ID, &h.Name, &h.CreatedBy, &h.CreatedAt, &h.UpdatedAt,
	); err != nil {
		return model.Household{}, fmt.Errorf("create household: %w", err)
	}

	const addOwner = `
		INSERT INTO household_members (user_id, household_id, role) VALUES ($1, $2, 'owner')
	`
	if _, err := tx.Exec(ctx, addOwner, createdByUserID, h.ID); err != nil {
		return model.Household{}, fmt.Errorf("add owner membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Household{}, fmt.Errorf("commit create household: %w", err)
	}

	return h, nil
}

func (r *HouseholdRepository) GetByID(ctx context.Context, householdID string) (model.Household, error) {
	const query = `
		SELECT id, name, created_by, created_at, updated_at
		FROM households WHERE id = $1
	`

	var h model.Household
	err := r.pool.QueryRow(ctx, query, householdID).Scan(
		&h.ID, &h.Name, &h.CreatedBy, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Household{}, ErrHouseholdNotFound
		}
		return model.Household{}, fmt.Errorf("get household: %w", err)
	}

	return h, nil
}

func (r *HouseholdRepository) Update(ctx context.Context, householdID, name string) (model.Household, error) {
	const query = `
		UPDATE households SET name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, created_by, created_at, updated_at
	`

	var h model.Household
	err := r.pool.QueryRow(ctx, query, name, householdID).Scan(
		&h.ID, &h.Name, &h.CreatedBy, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Household{}, ErrHouseholdNotFound
		}
		return model.Household{}, fmt.Errorf("update household: %w", err)
	}

	return h, nil
}

func (r *HouseholdRepository) Delete(ctx context.Context, householdID string) error {
	const query = `DELETE FROM households WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, householdID)
	if err != nil {
		return fmt.Errorf("delete household: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrHouseholdNotFound
	}

	return nil
}

// ── Members ───────────────────────────────────────────────────────────────────

func (r *HouseholdRepository) GetMembership(ctx context.Context, householdID, userID string) (model.HouseholdMember, error) {
	const query = `
		SELECT hm.id, hm.household_id, hm.user_id, u.display_name, u.email, hm.role::text, hm.joined_at
		FROM household_members hm
		JOIN users u ON u.id = hm.user_id
		WHERE hm.household_id = $1 AND hm.user_id = $2
	`

	var m model.HouseholdMember
	err := r.pool.QueryRow(ctx, query, householdID, userID).Scan(
		&m.ID, &m.HouseholdID, &m.UserID, &m.DisplayName, &m.Email, &m.Role, &m.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.HouseholdMember{}, ErrMemberNotFound
		}
		return model.HouseholdMember{}, fmt.Errorf("get membership: %w", err)
	}

	return m, nil
}

func (r *HouseholdRepository) ListMembers(ctx context.Context, householdID string) ([]model.HouseholdMember, error) {
	const query = `
		SELECT hm.id, hm.household_id, hm.user_id, u.display_name, u.email, hm.role::text, hm.joined_at
		FROM household_members hm
		JOIN users u ON u.id = hm.user_id
		WHERE hm.household_id = $1
		ORDER BY hm.joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, householdID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	members := make([]model.HouseholdMember, 0)
	for rows.Next() {
		var m model.HouseholdMember
		if err := rows.Scan(
			&m.ID, &m.HouseholdID, &m.UserID, &m.DisplayName, &m.Email, &m.Role, &m.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return members, nil
}

func (r *HouseholdRepository) UpdateMemberRole(ctx context.Context, householdID, userID, role string) (model.HouseholdMember, error) {
	const query = `
		UPDATE household_members SET role = $1::household_role
		WHERE household_id = $2 AND user_id = $3
		RETURNING id, household_id, user_id, role::text, joined_at
	`

	// We need the user's display_name and email too; fetch after update.
	var m model.HouseholdMember
	err := r.pool.QueryRow(ctx, query, role, householdID, userID).Scan(
		&m.ID, &m.HouseholdID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.HouseholdMember{}, ErrMemberNotFound
		}
		return model.HouseholdMember{}, fmt.Errorf("update member role: %w", err)
	}

	// Fetch user details.
	const userQuery = `SELECT display_name, email FROM users WHERE id = $1`
	if err := r.pool.QueryRow(ctx, userQuery, m.UserID).Scan(&m.DisplayName, &m.Email); err != nil {
		return model.HouseholdMember{}, fmt.Errorf("fetch user for member: %w", err)
	}

	return m, nil
}

func (r *HouseholdRepository) RemoveMember(ctx context.Context, householdID, userID string) error {
	const query = `DELETE FROM household_members WHERE household_id = $1 AND user_id = $2`

	tag, err := r.pool.Exec(ctx, query, householdID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *HouseholdRepository) AddMember(ctx context.Context, householdID, userID, role string) (model.HouseholdMember, error) {
	const query = `
		INSERT INTO household_members (user_id, household_id, role)
		VALUES ($1, $2, $3::household_role)
		RETURNING id, household_id, user_id, role::text, joined_at
	`

	var m model.HouseholdMember
	err := r.pool.QueryRow(ctx, query, userID, householdID, role).Scan(
		&m.ID, &m.HouseholdID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.HouseholdMember{}, ErrAlreadyMember
		}
		return model.HouseholdMember{}, fmt.Errorf("add member: %w", err)
	}

	const userQuery = `SELECT display_name, email FROM users WHERE id = $1`
	if err := r.pool.QueryRow(ctx, userQuery, userID).Scan(&m.DisplayName, &m.Email); err != nil {
		return model.HouseholdMember{}, fmt.Errorf("fetch user for member: %w", err)
	}

	return m, nil
}

func (r *HouseholdRepository) CountOwners(ctx context.Context, householdID string) (int, error) {
	const query = `
		SELECT COUNT(*) FROM household_members
		WHERE household_id = $1 AND role = 'owner'
	`

	var count int
	if err := r.pool.QueryRow(ctx, query, householdID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count owners: %w", err)
	}

	return count, nil
}

// ── Invitations ───────────────────────────────────────────────────────────────

type CreateInvitationParams struct {
	HouseholdID    string
	Email          string
	Role           string
	TokenHash      string
	ExpiresAt      time.Time
}

func (r *HouseholdRepository) CreateInvitation(ctx context.Context, params CreateInvitationParams) (model.Invitation, error) {
	const query = `
		INSERT INTO invitations (household_id, email, role, token_hash, expires_at)
		VALUES ($1, $2, $3::household_role, $4, $5)
		RETURNING id, household_id, email, role::text, status::text, expires_at, created_at
	`

	var inv model.Invitation
	err := r.pool.QueryRow(ctx, query,
		params.HouseholdID,
		params.Email,
		params.Role,
		params.TokenHash,
		params.ExpiresAt,
	).Scan(
		&inv.ID, &inv.HouseholdID, &inv.Email, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Invitation{}, ErrInvitationConflict
		}
		return model.Invitation{}, fmt.Errorf("create invitation: %w", err)
	}

	return inv, nil
}

func (r *HouseholdRepository) ListPendingInvitations(ctx context.Context, householdID string) ([]model.Invitation, error) {
	const query = `
		SELECT id, household_id, email, role::text, status::text, expires_at, created_at
		FROM invitations
		WHERE household_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, householdID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	invitations := make([]model.Invitation, 0)
	for rows.Next() {
		var inv model.Invitation
		if err := rows.Scan(
			&inv.ID, &inv.HouseholdID, &inv.Email, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invitations: %w", err)
	}

	return invitations, nil
}

func (r *HouseholdRepository) UpdateInvitationStatus(ctx context.Context, invitationID, householdID, status string) error {
	const query = `
		UPDATE invitations SET status = $1::invitation_status
		WHERE id = $2 AND household_id = $3 AND status = 'pending'
	`

	tag, err := r.pool.Exec(ctx, query, status, invitationID, householdID)
	if err != nil {
		return fmt.Errorf("update invitation status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}

	return nil
}

func (r *HouseholdRepository) FindPendingInvitationByTokenHash(ctx context.Context, tokenHash string) (model.Invitation, error) {
	const query = `
		SELECT id, household_id, email, role::text, status::text, expires_at, created_at
		FROM invitations
		WHERE token_hash = $1 AND status = 'pending'
	`

	var inv model.Invitation
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&inv.ID, &inv.HouseholdID, &inv.Email, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Invitation{}, ErrInvitationNotFound
		}
		return model.Invitation{}, fmt.Errorf("find invitation by token: %w", err)
	}

	return inv, nil
}

// AcceptInvitation atomically marks an invitation as accepted and adds the user as a member.
func (r *HouseholdRepository) AcceptInvitation(ctx context.Context, invitationID, householdID, userID, role string) (model.HouseholdMember, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.HouseholdMember{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const markAccepted = `
		UPDATE invitations SET status = 'accepted'
		WHERE id = $1 AND household_id = $2 AND status = 'pending'
	`
	tag, err := tx.Exec(ctx, markAccepted, invitationID, householdID)
	if err != nil {
		return model.HouseholdMember{}, fmt.Errorf("mark invitation accepted: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.HouseholdMember{}, ErrInvitationNotFound
	}

	const addMember = `
		INSERT INTO household_members (user_id, household_id, role)
		VALUES ($1, $2, $3::household_role)
		RETURNING id, household_id, user_id, role::text, joined_at
	`
	var m model.HouseholdMember
	err = tx.QueryRow(ctx, addMember, userID, householdID, role).Scan(
		&m.ID, &m.HouseholdID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.HouseholdMember{}, ErrAlreadyMember
		}
		return model.HouseholdMember{}, fmt.Errorf("add member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.HouseholdMember{}, fmt.Errorf("commit accept invitation: %w", err)
	}

	const userQuery = `SELECT display_name, email FROM users WHERE id = $1`
	if err := r.pool.QueryRow(ctx, userQuery, userID).Scan(&m.DisplayName, &m.Email); err != nil {
		return model.HouseholdMember{}, fmt.Errorf("fetch user for member: %w", err)
	}

	return m, nil
}
