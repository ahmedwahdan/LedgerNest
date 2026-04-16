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

type categoryService interface {
	List(ctx context.Context, requesterID string, householdID *string) ([]model.Category, error)
	Create(ctx context.Context, requesterID, householdID, name string, parentID, icon, color *string) (model.Category, error)
	Update(ctx context.Context, requesterID, categoryID, householdID, name string, parentID, icon, color *string) (model.Category, error)
	Delete(ctx context.Context, requesterID, categoryID, householdID string) error
}

type CategoryHandler struct {
	categories categoryService
}

func NewCategoryHandler(categories categoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

// List handles GET /categories
// Optional query param: household_id — includes household-specific categories alongside system ones.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var householdID *string
	if hid := r.URL.Query().Get("household_id"); hid != "" {
		householdID = &hid
	}

	cats, err := h.categories.List(r.Context(), claims.UserID, householdID)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"categories": cats})
}

type createCategoryRequest struct {
	HouseholdID string  `json:"household_id"`
	Name        string  `json:"name"`
	ParentID    *string `json:"parent_id"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
}

// Create handles POST /categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.HouseholdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id is required")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	cat, err := h.categories.Create(r.Context(), claims.UserID, req.HouseholdID, req.Name, req.ParentID, req.Icon, req.Color)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"category": cat})
}

type updateCategoryRequest struct {
	HouseholdID string  `json:"household_id"`
	Name        string  `json:"name"`
	ParentID    *string `json:"parent_id"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
}

// Update handles PUT /categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	categoryID := r.PathValue("id")
	if categoryID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing category id")
		return
	}

	var req updateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.HouseholdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id is required")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	cat, err := h.categories.Update(r.Context(), claims.UserID, categoryID, req.HouseholdID, req.Name, req.ParentID, req.Icon, req.Color)
	if err != nil {
		writeCategoryError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"category": cat})
}

// Delete handles DELETE /categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	categoryID := r.PathValue("id")
	if categoryID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing category id")
		return
	}

	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "household_id query param is required")
		return
	}

	if err := h.categories.Delete(r.Context(), claims.UserID, categoryID, householdID); err != nil {
		writeCategoryError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeCategoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCategoryNotFound):
		httpx.WriteError(w, http.StatusNotFound, "category not found")
	case errors.Is(err, service.ErrCategoryNameConflict):
		httpx.WriteError(w, http.StatusConflict, "category name already exists in this household")
	case errors.Is(err, service.ErrNotMember):
		httpx.WriteError(w, http.StatusForbidden, "you are not a member of this household")
	case errors.Is(err, service.ErrInsufficientRole):
		httpx.WriteError(w, http.StatusForbidden, "you don't have permission to perform this action")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "failed to process category request")
	}
}
