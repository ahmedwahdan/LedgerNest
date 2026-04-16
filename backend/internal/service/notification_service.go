package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var thresholds = []int{50, 75, 90, 100}

type notificationStore interface {
	Create(ctx context.Context, p repository.WriteNotificationParams) (model.Notification, error)
	List(ctx context.Context, userID string, limit int) ([]model.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, notificationID, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	ExistsForBudgetThreshold(ctx context.Context, userID, budgetID, snapshotID string, threshold int) (bool, error)
}

type NotificationService struct {
	store notificationStore
}

func NewNotificationService(store notificationStore) *NotificationService {
	return &NotificationService{store: store}
}

func (s *NotificationService) List(ctx context.Context, userID string, limit int) ([]model.Notification, error) {
	notifications, err := s.store.List(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	if notifications == nil {
		notifications = []model.Notification{}
	}
	return notifications, nil
}

func (s *NotificationService) CountUnread(ctx context.Context, userID string) (int, error) {
	return s.store.CountUnread(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, notificationID, userID string) error {
	return s.store.MarkRead(ctx, notificationID, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.store.MarkAllRead(ctx, userID)
}

// CheckBudgetThresholds is called after computing budget health. It creates
// threshold notifications for any budget that has crossed a new watermark
// (50 / 75 / 90 / 100 %) since the last check. Failures are logged and
// swallowed so the health response is never blocked.
func (s *NotificationService) CheckBudgetThresholds(ctx context.Context, userID string, items []model.BudgetHealthItem, snapshotID string) {
	for _, item := range items {
		for _, t := range thresholds {
			if item.PctUsed < float64(t) {
				continue
			}

			exists, err := s.store.ExistsForBudgetThreshold(ctx, userID, item.BudgetID, snapshotID, t)
			if err != nil {
				slog.Warn("check budget threshold exists", "error", err)
				continue
			}
			if exists {
				continue
			}

			label := "overall"
			if item.CategoryName != nil {
				label = *item.CategoryName
			}

			title := fmt.Sprintf("%d%% of %s budget used", t, label)
			body := fmt.Sprintf(
				"You've used %.0f%% (%.2f / %.2f) of your %s budget for this cycle.",
				item.PctUsed, mustFloat(item.Spent), mustFloat(item.Amount), label,
			)

			if _, err := s.store.Create(ctx, repository.WriteNotificationParams{
				UserID: userID,
				Type:   "budget_threshold",
				Title:  title,
				Body:   body,
				Metadata: map[string]any{
					"budget_id":   item.BudgetID,
					"snapshot_id": snapshotID,
					"threshold":   strconv.Itoa(t),
					"category":    label,
				},
			}); err != nil {
				slog.Warn("create budget threshold notification", "error", err)
			}
		}
	}
}

func mustFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
