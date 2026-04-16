package model

import "time"

// BudgetCycleConfig defines when a household's billing cycle starts.
// The active config is the most recently created one.
type BudgetCycleConfig struct {
	ID            string    `json:"id"`
	HouseholdID   string    `json:"household_id"`
	StartDay      int       `json:"start_day"`
	EffectiveFrom string    `json:"effective_from"` // DATE as "YYYY-MM-DD"
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// CycleSnapshot represents a single billing period for a household.
type CycleSnapshot struct {
	ID          string    `json:"id"`
	HouseholdID string    `json:"household_id"`
	CycleStart  string    `json:"cycle_start"` // DATE as "YYYY-MM-DD"
	CycleEnd    string    `json:"cycle_end"`
	Label       string    `json:"label"`
	Status      string    `json:"status"` // open | closed
	ConfigID    string    `json:"config_id"`
	CreatedAt   time.Time `json:"created_at"`
}
