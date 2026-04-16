package service

import (
	"context"
	"log/slog"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

type auditWriter interface {
	Write(ctx context.Context, params repository.WriteAuditParams) error
	ListByEntity(ctx context.Context, entityType, entityID string) ([]model.AuditLogEntry, error)
}

// AuditService writes audit log entries and retrieves history.
// Write failures are logged but never returned to callers — audit logging is
// best-effort and must not block the primary operation.
type AuditService struct {
	log auditWriter
}

func NewAuditService(log auditWriter) *AuditService {
	return &AuditService{log: log}
}

func (s *AuditService) Record(ctx context.Context, params repository.WriteAuditParams) {
	if err := s.log.Write(ctx, params); err != nil {
		slog.Warn("audit log write failed", "error", err,
			"action", params.Action,
			"entity_type", params.EntityType,
			"entity_id", params.EntityID,
		)
	}
}

func (s *AuditService) ListByEntity(ctx context.Context, entityType, entityID string) ([]model.AuditLogEntry, error) {
	return s.log.ListByEntity(ctx, entityType, entityID)
}
