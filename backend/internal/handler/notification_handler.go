package handler

import (
	"context"
	"net/http"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
)

type notificationService interface {
	List(ctx context.Context, userID string, limit int) ([]model.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, notificationID, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
}

type NotificationHandler struct {
	notifications notificationService
}

func NewNotificationHandler(notifications notificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

// List handles GET /notifications
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	notifications, err := h.notifications.List(r.Context(), claims.UserID, queryInt(r, "limit", 30))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}

	unread, _ := h.notifications.CountUnread(r.Context(), claims.UserID)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
		"unread_count":  unread,
	})
}

// MarkRead handles PUT /notifications/:id/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	if err := h.notifications.MarkRead(r.Context(), r.PathValue("id"), claims.UserID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// MarkAllRead handles PUT /notifications/read-all
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	if err := h.notifications.MarkAllRead(r.Context(), claims.UserID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to mark all notifications as read")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
