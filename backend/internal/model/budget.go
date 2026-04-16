package model

import "time"

// Budget defines a spending limit for a category (or overall) within a cycle snapshot.
type Budget struct {
	ID              string    `json:"id"`
	Scope           string    `json:"scope"` // personal | household
	UserID          *string   `json:"user_id,omitempty"`
	HouseholdID     *string   `json:"household_id,omitempty"`
	CategoryID      *string   `json:"category_id,omitempty"` // null = overall cap
	CycleSnapshotID string    `json:"cycle_snapshot_id"`
	Amount          string    `json:"amount"`
	RolloverAmount  string    `json:"rollover_amount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BudgetHealthItem is one row in the health summary.
type BudgetHealthItem struct {
	BudgetID     string  `json:"budget_id"`
	CategoryID   *string `json:"category_id,omitempty"`
	CategoryName *string `json:"category_name,omitempty"`
	Amount       string  `json:"amount"`
	Rollover     string  `json:"rollover"`
	Spent        string  `json:"spent"`
	Remaining    string  `json:"remaining"`
	PctUsed      float64 `json:"pct_used"`
}

// BudgetHealth is the full health response for a cycle snapshot.
type BudgetHealth struct {
	Snapshot   CycleSnapshot      `json:"snapshot"`
	Overall    *BudgetHealthItem  `json:"overall,omitempty"`
	Categories []BudgetHealthItem `json:"categories"`
}
