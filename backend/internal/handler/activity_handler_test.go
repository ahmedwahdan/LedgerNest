package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

func TestActivityHandlerListScopesToCurrentUser(t *testing.T) {
	t.Parallel()

	stub := &activityServiceStub{}
	handler := NewActivityHandler(stub)

	request := httptest.NewRequest(http.MethodGet, "/activity", http.NoBody)
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if stub.filters.UserID == nil || *stub.filters.UserID != "user-1" {
		t.Fatalf("expected user filter to be scoped to current user, got %#v", stub.filters.UserID)
	}
}

func TestActivityHandlerListRejectsAnotherUser(t *testing.T) {
	t.Parallel()

	handler := NewActivityHandler(&activityServiceStub{})

	request := httptest.NewRequest(http.MethodGet, "/activity?user_id=user-2", http.NoBody)
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
}

type activityServiceStub struct {
	filters repository.ActivityFilters
}

func (s *activityServiceStub) ListActivity(_ context.Context, f repository.ActivityFilters) ([]model.AuditLogEntry, error) {
	s.filters = f
	return []model.AuditLogEntry{}, nil
}
