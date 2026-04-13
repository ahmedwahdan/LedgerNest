package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

type budgetCycleService interface {
	GetCycle(ctx context.Context, requesterID, householdID string) (service.CycleState, error)
	SetCycleConfig(ctx context.Context, requesterID, householdID string, startDay int) (service.CycleState, error)
	ListSnapshots(ctx context.Context, requesterID, householdID string) ([]model.CycleSnapshot, error)
}

type BudgetCycleHandler struct {
	cycles budgetCycleService
}

func NewBudgetCycleHandler(cycles budgetCycleService) *BudgetCycleHandler {
	return &BudgetCycleHandler{cycles: cycles}
}

// GetCycle handles GET /households/{id}/cycle
func (h *BudgetCycleHandler) GetCycle(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	state, err := h.cycles.GetCycle(r.Context(), claims.UserID, householdID)
	if err != nil {
		writeCycleError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, state)
}

// SetCycleConfig handles PUT /households/{id}/cycle
func (h *BudgetCycleHandler) SetCycleConfig(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	var req struct {
		StartDay int `json:"start_day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.StartDay == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "start_day is required")
		return
	}

	state, err := h.cycles.SetCycleConfig(r.Context(), claims.UserID, householdID, req.StartDay)
	if err != nil {
		writeCycleError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, state)
}

// ListSnapshots handles GET /households/{id}/cycle/snapshots
func (h *BudgetCycleHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.PathValue("id")

	snapshots, err := h.cycles.ListSnapshots(r.Context(), claims.UserID, householdID)
	if err != nil {
		writeCycleError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"snapshots": snapshots})
}

func writeCycleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCycleConfigNotFound):
		httpx.WriteError(w, http.StatusNotFound, "no cycle config found — set one with PUT /households/:id/cycle")
	case errors.Is(err, service.ErrCycleSnapshotNotFound):
		httpx.WriteError(w, http.StatusNotFound, "cycle snapshot not found")
	case errors.Is(err, service.ErrNotMember):
		httpx.WriteError(w, http.StatusForbidden, "you are not a member of this household")
	case errors.Is(err, service.ErrInsufficientRole):
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to perform this action")
	case errors.Is(err, service.ErrInvalidStartDay):
		httpx.WriteError(w, http.StatusBadRequest, "start_day must be between 1 and 28")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
