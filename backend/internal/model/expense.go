package model

import "time"

type Expense struct {
	ID                 string    `json:"id"`
	Scope              string    `json:"scope"`
	UserID             string    `json:"user_id"`
	CreatedBy          string    `json:"created_by"`
	Amount             string    `json:"amount"`
	Currency           string    `json:"currency"`
	Merchant           string    `json:"merchant"`
	CategoryID         *string   `json:"category_id,omitempty"`
	PaymentMethod      string    `json:"payment_method"`
	Date               string    `json:"date"`
	Notes              string    `json:"notes,omitempty"`
	IsRecurring        bool      `json:"is_recurring"`
	RecurrenceInterval *string   `json:"recurrence_interval,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
