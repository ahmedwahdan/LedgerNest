package model

import "time"

// Household is a shared expense group.
type Household struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HouseholdMember links a user to a household with a role.
type HouseholdMember struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"household_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"` // owner | editor | viewer
	JoinedAt    time.Time `json:"joined_at"`
}

// Invitation represents a pending household invite sent by email.
type Invitation struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"household_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Status      string    `json:"status"` // pending | accepted | revoked
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}
