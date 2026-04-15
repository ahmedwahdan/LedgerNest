package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

type analyticsService interface {
	GetSpending(ctx context.Context, input service.AnalyticsInput) (
		summary repository.SpendingSummary,
		byCategory []repository.SpendingByCategory,
		err error,
	)
	GetTrends(ctx context.Context, input service.AnalyticsInput) ([]repository.MonthlyTrend, error)
	GetTopMerchants(ctx context.Context, input service.AnalyticsInput) ([]repository.TopMerchant, error)
}

type AnalyticsHandler struct {
	analytics analyticsService
}

func NewAnalyticsHandler(analytics analyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

// buildInput extracts common query params and returns an AnalyticsInput.
func (h *AnalyticsHandler) buildInput(r *http.Request, userID string) service.AnalyticsInput {
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "personal"
	}
	return service.AnalyticsInput{
		Scope:       scope,
		UserID:      userID,
		HouseholdID: r.URL.Query().Get("household_id"),
		From:        r.URL.Query().Get("from"),
		To:          r.URL.Query().Get("to"),
	}
}

// Spending handles GET /analytics/spending
func (h *AnalyticsHandler) Spending(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	input := h.buildInput(r, claims.UserID)

	summary, byCategory, err := h.analytics.GetSpending(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrNotMember) {
			httpx.WriteError(w, http.StatusForbidden, "not a member of this household")
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"summary":     summary,
		"by_category": byCategory,
	})
}

// Trends handles GET /analytics/trends
func (h *AnalyticsHandler) Trends(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	input := h.buildInput(r, claims.UserID)

	trends, err := h.analytics.GetTrends(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrNotMember) {
			httpx.WriteError(w, http.StatusForbidden, "not a member of this household")
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"trends": trends})
}

// TopMerchants handles GET /analytics/merchants
func (h *AnalyticsHandler) TopMerchants(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	input := h.buildInput(r, claims.UserID)
	input.Limit = queryInt(r, "limit", 10)

	merchants, err := h.analytics.GetTopMerchants(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrNotMember) {
			httpx.WriteError(w, http.StatusForbidden, "not a member of this household")
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"merchants": merchants})
}
