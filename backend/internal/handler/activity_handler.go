package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

type activityService interface {
	ListActivity(ctx context.Context, f repository.ActivityFilters) ([]model.AuditLogEntry, error)
}

type ActivityHandler struct {
	audit activityService
}

func NewActivityHandler(audit activityService) *ActivityHandler {
	return &ActivityHandler{audit: audit}
}

// List handles GET /activity
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	f := repository.ActivityFilters{
		UserID: ptr(claims.UserID),
		Page:  queryInt(r, "page", 1),
		Limit: queryInt(r, "limit", 20),
	}

	if uid := r.URL.Query().Get("user_id"); uid != "" {
		if uid != claims.UserID {
			httpx.WriteError(w, http.StatusForbidden, "you can only view your own activity")
			return
		}
	}
	if et := r.URL.Query().Get("entity_type"); et != "" {
		f.EntityType = &et
	}
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "from must be YYYY-MM-DD")
			return
		}
		f.From = &t
	}
	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "to must be YYYY-MM-DD")
			return
		}
		end := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		f.To = &end
	}

	entries, err := h.audit.ListActivity(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list activity")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"activity": entries})
}

func ptr[T any](value T) *T {
	return &value
}
