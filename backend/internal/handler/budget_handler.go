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

type budgetService interface {
	List(ctx context.Context, requesterID, householdID string, snapshotID *string, scope *string) ([]model.Budget, error)
	Create(ctx context.Context, requesterID string, input service.CreateBudgetInput) (model.Budget, error)
	Update(ctx context.Context, requesterID, householdID, budgetID, amount string) (model.Budget, error)
	Delete(ctx context.Context, requesterID, householdID, budgetID string) error
	GetHealth(ctx context.Context, requesterID, householdID string, snapshotID *string, scope string) (model.BudgetHealth, error)
}

type BudgetHandler struct {
	budgets budgetService
}

func NewBudgetHandler(budgets budgetService) *BudgetHandler {
	return &BudgetHandler{budgets: budgets}
}

// List handles GET /budgets?household_id=&snapshot_id=&scope=
func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id query param is required")
		return
	}

	var snapshotID *string
	if sid := r.URL.Query().Get("snapshot_id"); sid != "" {
		snapshotID = &sid
	}

	var scope *string
	if sc := r.URL.Query().Get("scope"); sc != "" {
		scope = &sc
	}

	budgets, err := h.budgets.List(r.Context(), claims.UserID, householdID, snapshotID, scope)
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"budgets": budgets})
}

// GetHealth handles GET /budgets/health?household_id=&scope=&snapshot_id=
func (h *BudgetHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id query param is required")
		return
	}

	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "household"
	}

	var snapshotID *string
	if sid := r.URL.Query().Get("snapshot_id"); sid != "" {
		snapshotID = &sid
	}

	health, err := h.budgets.GetHealth(r.Context(), claims.UserID, householdID, snapshotID, scope)
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, health)
}

type createBudgetRequest struct {
	HouseholdID string  `json:"household_id"`
	Scope       string  `json:"scope"`
	CategoryID  *string `json:"category_id"`
	SnapshotID  *string `json:"snapshot_id"`
	Amount      string  `json:"amount"`
}

// Create handles POST /budgets
func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	var req createBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	switch {
	case req.HouseholdID == "":
		httpx.WriteError(w, http.StatusBadRequest, "household_id is required")
		return
	case req.Scope == "":
		httpx.WriteError(w, http.StatusBadRequest, "scope is required")
		return
	case req.Amount == "":
		httpx.WriteError(w, http.StatusBadRequest, "amount is required")
		return
	}

	budget, err := h.budgets.Create(r.Context(), claims.UserID, service.CreateBudgetInput{
		HouseholdID: req.HouseholdID,
		Scope:       req.Scope,
		CategoryID:  req.CategoryID,
		SnapshotID:  req.SnapshotID,
		Amount:      req.Amount,
	})
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"budget": budget})
}

// Update handles PUT /budgets/{id}
func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	budgetID := r.PathValue("id")

	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id query param is required")
		return
	}

	var req struct {
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Amount == "" {
		httpx.WriteError(w, http.StatusBadRequest, "amount is required")
		return
	}

	budget, err := h.budgets.Update(r.Context(), claims.UserID, householdID, budgetID, req.Amount)
	if err != nil {
		writeBudgetError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"budget": budget})
}

// Delete handles DELETE /budgets/{id}
func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)
	budgetID := r.PathValue("id")

	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id query param is required")
		return
	}

	if err := h.budgets.Delete(r.Context(), claims.UserID, householdID, budgetID); err != nil {
		writeBudgetError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeBudgetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBudgetNotFound):
		httpx.WriteError(w, http.StatusNotFound, "budget not found")
	case errors.Is(err, service.ErrBudgetConflict):
		httpx.WriteError(w, http.StatusConflict, "a budget for this category already exists in this snapshot")
	case errors.Is(err, service.ErrBudgetForbidden):
		httpx.WriteError(w, http.StatusForbidden, "you don't have access to this budget")
	case errors.Is(err, service.ErrNotMember):
		httpx.WriteError(w, http.StatusForbidden, "you are not a member of this household")
	case errors.Is(err, service.ErrCycleSnapshotNotFound):
		httpx.WriteError(w, http.StatusNotFound, "cycle snapshot not found")
	default:
		msg := err.Error()
		if msg != "" && len(msg) < 120 {
			httpx.WriteError(w, http.StatusBadRequest, msg)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
