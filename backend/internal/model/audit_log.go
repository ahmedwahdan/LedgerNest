package model

import "time"

// AuditLogEntry records a single create/update/delete/restore action on a tracked entity.
type AuditLogEntry struct {
	ID         string         `json:"id"`
	UserID     *string        `json:"user_id,omitempty"`
	Action     string         `json:"action"` // create | update | delete | restore
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	OldValues  map[string]any `json:"old_values,omitempty"`
	NewValues  map[string]any `json:"new_values,omitempty"`
	IPAddress  *string        `json:"ip_address,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}
