package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var ErrInvalidExpenseInput = errors.New("invalid expense input")

type expenseStore interface {
	CreatePersonal(ctx context.Context, params repository.CreateExpenseParams) (model.Expense, error)
	ListPersonalByUserID(ctx context.Context, userID string) ([]model.Expense, error)
	ListPersonal(ctx context.Context, filters repository.ListExpenseFilters) ([]model.Expense, error)
	GetPersonalByID(ctx context.Context, expenseID, userID string) (model.Expense, error)
	UpdatePersonal(ctx context.Context, params repository.UpdateExpenseParams) (model.Expense, error)
	DeletePersonal(ctx context.Context, expenseID, userID, deletedBy string) error
	RestorePersonal(ctx context.Context, expenseID, userID, restoredBy string) (model.Expense, error)
}

type CreateExpenseInput struct {
	Amount             string
	Currency           string
	Merchant           string
	CategoryID         *string
	PaymentMethod      string
	Date               string
	Notes              string
	IsRecurring        bool
	RecurrenceInterval *string
}

type ListExpensesInput struct {
	From       string
	To         string
	Merchant   string
	CategoryID *string
	Limit      int
	Offset     int
}

type auditRecorder interface {
	Record(ctx context.Context, params repository.WriteAuditParams)
}

type ExpenseService struct {
	expenses expenseStore
	audit    auditRecorder
}

func NewExpenseService(expenses expenseStore, audit auditRecorder) *ExpenseService {
	return &ExpenseService{expenses: expenses, audit: audit}
}

func (s *ExpenseService) CreatePersonal(ctx context.Context, userID string, input CreateExpenseInput) (model.Expense, error) {
	normalized, err := normalizeExpenseInput(input)
	if err != nil {
		return model.Expense{}, err
	}

	expense, err := s.expenses.CreatePersonal(ctx, repository.CreateExpenseParams{
		UserID:             userID,
		CreatedBy:          userID,
		Amount:             normalized.Amount,
		Currency:           normalized.Currency,
		Merchant:           normalized.Merchant,
		CategoryID:         normalized.CategoryID,
		PaymentMethod:      normalized.PaymentMethod,
		Date:               normalized.Date,
		Notes:              normalized.Notes,
		IsRecurring:        normalized.IsRecurring,
		RecurrenceInterval: normalized.RecurrenceInterval,
	})
	if err != nil {
		return model.Expense{}, err
	}

	s.audit.Record(ctx, repository.WriteAuditParams{
		UserID:     &userID,
		Action:     "create",
		EntityType: "expense",
		EntityID:   expense.ID,
		NewValues:  expense,
	})

	return expense, nil
}

func (s *ExpenseService) ListPersonal(ctx context.Context, userID string, input ListExpensesInput) ([]model.Expense, error) {
	filters, err := normalizeListExpensesInput(input)
	if err != nil {
		return nil, err
	}

	return s.expenses.ListPersonal(ctx, repository.ListExpenseFilters{
		UserID:     userID,
		From:       filters.From,
		To:         filters.To,
		Merchant:   filters.Merchant,
		CategoryID: filters.CategoryID,
		Limit:      input.Limit,
		Offset:     input.Offset,
	})
}

func (s *ExpenseService) GetPersonal(ctx context.Context, expenseID, userID string) (model.Expense, error) {
	if strings.TrimSpace(expenseID) == "" {
		return model.Expense{}, fmt.Errorf("%w: expense id is required", ErrInvalidExpenseInput)
	}

	return s.expenses.GetPersonalByID(ctx, expenseID, userID)
}

func (s *ExpenseService) UpdatePersonal(ctx context.Context, expenseID, userID string, input CreateExpenseInput) (model.Expense, error) {
	if strings.TrimSpace(expenseID) == "" {
		return model.Expense{}, fmt.Errorf("%w: expense id is required", ErrInvalidExpenseInput)
	}

	normalized, err := normalizeExpenseInput(input)
	if err != nil {
		return model.Expense{}, err
	}

	before, _ := s.expenses.GetPersonalByID(ctx, expenseID, userID)

	expense, err := s.expenses.UpdatePersonal(ctx, repository.UpdateExpenseParams{
		ExpenseID:          expenseID,
		UserID:             userID,
		UpdatedBy:          userID,
		Amount:             normalized.Amount,
		Currency:           normalized.Currency,
		Merchant:           normalized.Merchant,
		CategoryID:         normalized.CategoryID,
		PaymentMethod:      normalized.PaymentMethod,
		Date:               normalized.Date,
		Notes:              normalized.Notes,
		IsRecurring:        normalized.IsRecurring,
		RecurrenceInterval: normalized.RecurrenceInterval,
	})
	if err != nil {
		return model.Expense{}, err
	}

	s.audit.Record(ctx, repository.WriteAuditParams{
		UserID:     &userID,
		Action:     "update",
		EntityType: "expense",
		EntityID:   expense.ID,
		OldValues:  before,
		NewValues:  expense,
	})

	return expense, nil
}

func (s *ExpenseService) DeletePersonal(ctx context.Context, expenseID, userID string) error {
	if strings.TrimSpace(expenseID) == "" {
		return fmt.Errorf("%w: expense id is required", ErrInvalidExpenseInput)
	}

	if err := s.expenses.DeletePersonal(ctx, expenseID, userID, userID); err != nil {
		return err
	}

	s.audit.Record(ctx, repository.WriteAuditParams{
		UserID:     &userID,
		Action:     "delete",
		EntityType: "expense",
		EntityID:   expenseID,
	})

	return nil
}

func (s *ExpenseService) RestorePersonal(ctx context.Context, expenseID, userID string) (model.Expense, error) {
	if strings.TrimSpace(expenseID) == "" {
		return model.Expense{}, fmt.Errorf("%w: expense id is required", ErrInvalidExpenseInput)
	}

	expense, err := s.expenses.RestorePersonal(ctx, expenseID, userID, userID)
	if err != nil {
		return model.Expense{}, err
	}

	s.audit.Record(ctx, repository.WriteAuditParams{
		UserID:     &userID,
		Action:     "restore",
		EntityType: "expense",
		EntityID:   expense.ID,
		NewValues:  expense,
	})

	return expense, nil
}

func normalizeExpenseInput(input CreateExpenseInput) (CreateExpenseInput, error) {
	input.Amount = strings.TrimSpace(input.Amount)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.Merchant = strings.TrimSpace(input.Merchant)
	input.PaymentMethod = strings.TrimSpace(input.PaymentMethod)
	input.Date = strings.TrimSpace(input.Date)
	input.Notes = strings.TrimSpace(input.Notes)

	switch {
	case input.Amount == "":
		return CreateExpenseInput{}, fmt.Errorf("%w: amount is required", ErrInvalidExpenseInput)
	case input.Currency == "":
		input.Currency = "USD"
	case len(input.Currency) != 3:
		return CreateExpenseInput{}, fmt.Errorf("%w: currency must be a 3-letter code", ErrInvalidExpenseInput)
	case input.Merchant == "":
		return CreateExpenseInput{}, fmt.Errorf("%w: merchant is required", ErrInvalidExpenseInput)
	case input.PaymentMethod == "":
		return CreateExpenseInput{}, fmt.Errorf("%w: payment_method is required", ErrInvalidExpenseInput)
	case input.Date == "":
		return CreateExpenseInput{}, fmt.Errorf("%w: date is required", ErrInvalidExpenseInput)
	}

	amount, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil || amount <= 0 {
		return CreateExpenseInput{}, fmt.Errorf("%w: amount must be a positive number", ErrInvalidExpenseInput)
	}

	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		return CreateExpenseInput{}, fmt.Errorf("%w: date must be in YYYY-MM-DD format", ErrInvalidExpenseInput)
	}

	if input.CategoryID != nil {
		value := strings.TrimSpace(*input.CategoryID)
		if value == "" {
			input.CategoryID = nil
		} else {
			input.CategoryID = &value
		}
	}

	if input.RecurrenceInterval != nil {
		value := strings.TrimSpace(*input.RecurrenceInterval)
		if value == "" {
			input.RecurrenceInterval = nil
		} else {
			input.RecurrenceInterval = &value
		}
	}

	if input.IsRecurring && input.RecurrenceInterval == nil {
		return CreateExpenseInput{}, fmt.Errorf("%w: recurrence_interval is required when is_recurring is true", ErrInvalidExpenseInput)
	}

	if !input.IsRecurring {
		input.RecurrenceInterval = nil
	}

	return input, nil
}

func normalizeListExpensesInput(input ListExpensesInput) (ListExpensesInput, error) {
	input.From = strings.TrimSpace(input.From)
	input.To = strings.TrimSpace(input.To)
	input.Merchant = strings.TrimSpace(input.Merchant)

	if input.From != "" {
		if _, err := time.Parse("2006-01-02", input.From); err != nil {
			return ListExpensesInput{}, fmt.Errorf("%w: from must be in YYYY-MM-DD format", ErrInvalidExpenseInput)
		}
	}
	if input.To != "" {
		if _, err := time.Parse("2006-01-02", input.To); err != nil {
			return ListExpensesInput{}, fmt.Errorf("%w: to must be in YYYY-MM-DD format", ErrInvalidExpenseInput)
		}
	}
	if input.From != "" && input.To != "" && input.From > input.To {
		return ListExpensesInput{}, fmt.Errorf("%w: from must be before or equal to to", ErrInvalidExpenseInput)
	}

	if input.CategoryID != nil {
		value := strings.TrimSpace(*input.CategoryID)
		if value == "" {
			input.CategoryID = nil
		} else {
			input.CategoryID = &value
		}
	}

	return input, nil
}
