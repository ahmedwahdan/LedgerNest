package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

type householdService interface {
	Create(ctx context.Context, requesterID, name string) (model.Household, error)
	Get(ctx context.Context, requesterID, householdID string) (model.Household, error)
	Update(ctx context.Context, requesterID, householdID, name string) (model.Household, error)
	Delete(ctx context.Context, requesterID, householdID string) error
	Leave(ctx context.Context, requesterID, householdID string) error

	ListMembers(ctx context.Context, requesterID, householdID string) ([]model.HouseholdMember, error)
	UpdateMemberRole(ctx context.Context, requesterID, householdID, targetUserID, role string) (model.HouseholdMember, error)
	RemoveMember(ctx context.Context, requesterID, householdID, targetUserID string) error

	CreateInvitation(ctx context.Context, requesterID, householdID, email, role string) (service.InviteResult, error)
	ListInvitations(ctx context.Context, requesterID, householdID string) ([]model.Invitation, error)
	RevokeInvitation(ctx context.Context, requesterID, householdID, invitationID string) error
	AcceptInvitation(ctx context.Context, userID, token string) (model.HouseholdMember, error)
}

type HouseholdHandler struct {
	households householdService
}

func NewHouseholdHandler(households householdService) *HouseholdHandler {
	return &HouseholdHandler{households: households}
}

// ── Household CRUD ────────────────────────────────────────────────────────────

// Create handles POST /households
func (h *HouseholdHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	household, err := h.households.Create(r.Context(), claims.UserID, req.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create household")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"household": household})
}

// Get handles GET /households/{id}
func (h *HouseholdHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	household, err := h.households.Get(r.Context(), claims.UserID, householdID)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"household": household})
}

// Update handles PUT /households/{id}
func (h *HouseholdHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	household, err := h.households.Update(r.Context(), claims.UserID, householdID, req.Name)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"household": household})
}

// Delete handles DELETE /households/{id}
func (h *HouseholdHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	if err := h.households.Delete(r.Context(), claims.UserID, householdID); err != nil {
		writeHouseholdError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Leave handles POST /households/{id}/leave
func (h *HouseholdHandler) Leave(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	if err := h.households.Leave(r.Context(), claims.UserID, householdID); err != nil {
		writeHouseholdError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Members ───────────────────────────────────────────────────────────────────

// ListMembers handles GET /households/{id}/members
func (h *HouseholdHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	members, err := h.households.ListMembers(r.Context(), claims.UserID, householdID)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"members": members})
}

// UpdateMemberRole handles PUT /households/{id}/members/{userId}/role
func (h *HouseholdHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")
	targetUserID := r.PathValue("userId")

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Role == "" {
		httpx.WriteError(w, http.StatusBadRequest, "role is required")
		return
	}

	member, err := h.households.UpdateMemberRole(r.Context(), claims.UserID, householdID, targetUserID, req.Role)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"member": member})
}

// RemoveMember handles DELETE /households/{id}/members/{userId}
func (h *HouseholdHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")
	targetUserID := r.PathValue("userId")

	if err := h.households.RemoveMember(r.Context(), claims.UserID, householdID, targetUserID); err != nil {
		writeHouseholdError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Invitations ───────────────────────────────────────────────────────────────

// CreateInvitation handles POST /households/{id}/invitations
func (h *HouseholdHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Role == "" {
		httpx.WriteError(w, http.StatusBadRequest, "role is required")
		return
	}

	result, err := h.households.CreateInvitation(r.Context(), claims.UserID, householdID, req.Email, req.Role)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

// ListInvitations handles GET /households/{id}/invitations
func (h *HouseholdHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	invitations, err := h.households.ListInvitations(r.Context(), claims.UserID, householdID)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"invitations": invitations})
}

// RevokeInvitation handles DELETE /households/{id}/invitations/{invId}
func (h *HouseholdHandler) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")
	invitationID := r.PathValue("invId")

	if err := h.households.RevokeInvitation(r.Context(), claims.UserID, householdID, invitationID); err != nil {
		writeHouseholdError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AcceptInvitation handles POST /invitations/accept
func (h *HouseholdHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Token == "" {
		httpx.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}

	member, err := h.households.AcceptInvitation(r.Context(), claims.UserID, req.Token)
	if err != nil {
		writeHouseholdError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"member": member})
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// mustClaims extracts access token claims; panics if absent (should only be called on auth-protected routes).
func mustClaims(r *http.Request) auth.AccessTokenClaims {
	claims, _ := auth.AccessTokenClaimsFromContext(r.Context())
	return claims
}

func writeHouseholdError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrHouseholdNotFound):
		httpx.WriteError(w, http.StatusNotFound, "household not found")
	case errors.Is(err, service.ErrNotMember):
		httpx.WriteError(w, http.StatusForbidden, "you are not a member of this household")
	case errors.Is(err, service.ErrInsufficientRole):
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to perform this action")
	case errors.Is(err, service.ErrMemberNotFound):
		httpx.WriteError(w, http.StatusNotFound, "member not found")
	case errors.Is(err, service.ErrAlreadyMember):
		httpx.WriteError(w, http.StatusConflict, "user is already a member of this household")
	case errors.Is(err, service.ErrCannotRemoveLastOwner):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "cannot remove the last owner of a household")
	case errors.Is(err, service.ErrInvitationNotFound):
		httpx.WriteError(w, http.StatusNotFound, "invitation not found")
	case errors.Is(err, service.ErrInvitationExpired):
		httpx.WriteError(w, http.StatusGone, "invitation has expired")
	case errors.Is(err, service.ErrInvitationConflict):
		httpx.WriteError(w, http.StatusConflict, "a pending invitation for this email already exists")
	default:
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		// surface validation messages (errors.New calls from service)
		if errMsg != "" && len(errMsg) < 100 {
			httpx.WriteError(w, http.StatusBadRequest, errMsg)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
