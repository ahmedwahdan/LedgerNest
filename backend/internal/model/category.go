package model

import "time"

// Category represents an expense classification, either system-defined or household-specific.
type Category struct {
	ID          string    `json:"id"`
	HouseholdID *string   `json:"household_id,omitempty"`
	Name        string    `json:"name"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Color       *string   `json:"color,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}
