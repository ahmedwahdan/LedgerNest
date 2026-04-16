package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

type expenseService interface {
	CreatePersonal(ctx context.Context, userID string, input service.CreateExpenseInput) (model.Expense, error)
	ListPersonal(ctx context.Context, userID string, input service.ListExpensesInput) ([]model.Expense, error)
	GetPersonal(ctx context.Context, expenseID, userID string) (model.Expense, error)
	UpdatePersonal(ctx context.Context, expenseID, userID string, input service.CreateExpenseInput) (model.Expense, error)
	DeletePersonal(ctx context.Context, expenseID, userID string) error
	RestorePersonal(ctx context.Context, expenseID, userID string) (model.Expense, error)
}

type auditService interface {
	ListByEntity(ctx context.Context, entityType, entityID string) ([]model.AuditLogEntry, error)
}

type ExpenseHandler struct {
	expenses expenseService
	audit    auditService
}

type createExpenseRequest struct {
	Amount             string  `json:"amount"`
	Currency           string  `json:"currency"`
	Merchant           string  `json:"merchant"`
	CategoryID         *string `json:"category_id"`
	PaymentMethod      string  `json:"payment_method"`
	Date               string  `json:"date"`
	Notes              string  `json:"notes"`
	IsRecurring        bool    `json:"is_recurring"`
	RecurrenceInterval *string `json:"recurrence_interval"`
}

func NewExpenseHandler(expenses expenseService, audit auditService) *ExpenseHandler {
	return &ExpenseHandler{expenses: expenses, audit: audit}
}

func (h *ExpenseHandler) CreatePersonal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	expense, err := h.expenses.CreatePersonal(r.Context(), claims.UserID, service.CreateExpenseInput{
		Amount:             req.Amount,
		Currency:           req.Currency,
		Merchant:           req.Merchant,
		CategoryID:         req.CategoryID,
		PaymentMethod:      req.PaymentMethod,
		Date:               req.Date,
		Notes:              req.Notes,
		IsRecurring:        req.IsRecurring,
		RecurrenceInterval: req.RecurrenceInterval,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpenseInput) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to create expense")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"expense": expense})
}

func (h *ExpenseHandler) ListPersonal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	expenses, err := h.expenses.ListPersonal(r.Context(), claims.UserID, service.ListExpensesInput{
		From:       r.URL.Query().Get("from"),
		To:         r.URL.Query().Get("to"),
		Merchant:   r.URL.Query().Get("merchant"),
		CategoryID: queryOptional(r, "category_id"),
		Limit:      queryInt(r, "limit", 50),
		Offset:     queryInt(r, "offset", 0),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpenseInput) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list expenses")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"expenses": expenses})
}

func (h *ExpenseHandler) GetPersonal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	expense, err := h.expenses.GetPersonal(r.Context(), r.PathValue("id"), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpenseInput) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, repository.ErrExpenseNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "expense not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load expense")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"expense": expense})
}

func (h *ExpenseHandler) UpdatePersonal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	expense, err := h.expenses.UpdatePersonal(r.Context(), r.PathValue("id"), claims.UserID, service.CreateExpenseInput{
		Amount:             req.Amount,
		Currency:           req.Currency,
		Merchant:           req.Merchant,
		CategoryID:         req.CategoryID,
		PaymentMethod:      req.PaymentMethod,
		Date:               req.Date,
		Notes:              req.Notes,
		IsRecurring:        req.IsRecurring,
		RecurrenceInterval: req.RecurrenceInterval,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpenseInput) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, repository.ErrExpenseNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "expense not found")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to update expense")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"expense": expense})
}

func (h *ExpenseHandler) DeletePersonal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	err := h.expenses.DeletePersonal(r.Context(), r.PathValue("id"), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpenseInput) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, repository.ErrExpenseNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "expense not found")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete expense")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ExpenseHandler) RestorePersonal(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	expense, err := h.expenses.RestorePersonal(r.Context(), r.PathValue("id"), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpenseInput) {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, repository.ErrExpenseNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "expense not found or not deleted")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to restore expense")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"expense": expense})
}

// GetHistory handles GET /expenses/{id}/history
func (h *ExpenseHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	expenseID := r.PathValue("id")

	// Verify ownership before exposing audit history.
	if _, err := h.expenses.GetPersonal(r.Context(), expenseID, claims.UserID); err != nil {
		if errors.Is(err, repository.ErrExpenseNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "expense not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to verify expense")
		return
	}

	entries, err := h.audit.ListByEntity(r.Context(), "expense", expenseID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load history")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"history": entries})
}

func queryOptional(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}

	return &value
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
