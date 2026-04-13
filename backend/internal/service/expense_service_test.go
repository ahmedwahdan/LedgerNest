package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

func TestExpenseServiceCreatePersonal(t *testing.T) {
	t.Parallel()

	store := &stubExpenseStore{
		createdExpense: model.Expense{
			ID:            "expense-1",
			UserID:        "user-1",
			CreatedBy:     "user-1",
			Amount:        "23.50",
			Currency:      "USD",
			Merchant:      "Market",
			PaymentMethod: "card",
			Date:          "2026-04-13",
		},
	}
	service := NewExpenseService(store, &stubAuditRecorder{})

	expense, err := service.CreatePersonal(context.Background(), "user-1", CreateExpenseInput{
		Amount:        "23.50",
		Currency:      "usd",
		Merchant:      "  Market  ",
		PaymentMethod: "card",
		Date:          "2026-04-13",
	})
	if err != nil {
		t.Fatalf("create personal returned error: %v", err)
	}

	if expense.ID != "expense-1" {
		t.Fatalf("unexpected expense id: %s", expense.ID)
	}
	if store.createParams.Currency != "USD" {
		t.Fatalf("expected normalized currency, got %s", store.createParams.Currency)
	}
	if store.createParams.Merchant != "Market" {
		t.Fatalf("expected trimmed merchant, got %q", store.createParams.Merchant)
	}
}

func TestExpenseServiceCreatePersonalRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewExpenseService(&stubExpenseStore{}, &stubAuditRecorder{})

	_, err := service.CreatePersonal(context.Background(), "user-1", CreateExpenseInput{
		Amount:        "-1",
		Merchant:      "Market",
		PaymentMethod: "card",
		Date:          "2026-04-13",
	})
	if !errors.Is(err, ErrInvalidExpenseInput) {
		t.Fatalf("expected invalid expense input, got %v", err)
	}
}

func TestExpenseServiceListPersonal(t *testing.T) {
	t.Parallel()

	store := &stubExpenseStore{
		listExpenses: []model.Expense{{ID: "expense-1"}, {ID: "expense-2"}},
	}
	service := NewExpenseService(store, &stubAuditRecorder{})

	categoryID := "category-1"
	expenses, err := service.ListPersonal(context.Background(), "user-1", ListExpensesInput{
		From:       "2026-04-01",
		To:         "2026-04-30",
		Merchant:   "market",
		CategoryID: &categoryID,
	})
	if err != nil {
		t.Fatalf("list personal returned error: %v", err)
	}
	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(expenses))
	}
	if store.listFilters.From != "2026-04-01" || store.listFilters.To != "2026-04-30" {
		t.Fatalf("unexpected list filters: %#v", store.listFilters)
	}
}

func TestExpenseServiceGetPersonal(t *testing.T) {
	t.Parallel()

	store := &stubExpenseStore{
		getExpense: model.Expense{ID: "expense-1"},
	}
	service := NewExpenseService(store, &stubAuditRecorder{})

	expense, err := service.GetPersonal(context.Background(), "expense-1", "user-1")
	if err != nil {
		t.Fatalf("get personal returned error: %v", err)
	}
	if expense.ID != "expense-1" {
		t.Fatalf("unexpected expense id: %s", expense.ID)
	}
	if store.getExpenseID != "expense-1" || store.getUserID != "user-1" {
		t.Fatalf("unexpected get args: expense=%s user=%s", store.getExpenseID, store.getUserID)
	}
}

func TestExpenseServiceUpdatePersonal(t *testing.T) {
	t.Parallel()

	store := &stubExpenseStore{
		updatedExpense: model.Expense{ID: "expense-1", Merchant: "Updated Market"},
	}
	service := NewExpenseService(store, &stubAuditRecorder{})

	expense, err := service.UpdatePersonal(context.Background(), "expense-1", "user-1", CreateExpenseInput{
		Amount:        "11.00",
		Currency:      "usd",
		Merchant:      " Updated Market ",
		PaymentMethod: "cash",
		Date:          "2026-04-13",
	})
	if err != nil {
		t.Fatalf("update personal returned error: %v", err)
	}
	if expense.ID != "expense-1" {
		t.Fatalf("unexpected expense id: %s", expense.ID)
	}
	if store.updateParams.ExpenseID != "expense-1" {
		t.Fatalf("unexpected update expense id: %s", store.updateParams.ExpenseID)
	}
}

func TestExpenseServiceDeletePersonal(t *testing.T) {
	t.Parallel()

	store := &stubExpenseStore{}
	service := NewExpenseService(store, &stubAuditRecorder{})

	if err := service.DeletePersonal(context.Background(), "expense-1", "user-1"); err != nil {
		t.Fatalf("delete personal returned error: %v", err)
	}
	if store.deletedExpenseID != "expense-1" {
		t.Fatalf("unexpected deleted expense id: %s", store.deletedExpenseID)
	}
}

type stubAuditRecorder struct{}

func (s *stubAuditRecorder) Record(context.Context, repository.WriteAuditParams) {}

type stubExpenseStore struct {
	createParams     repository.CreateExpenseParams
	updateParams     repository.UpdateExpenseParams
	createdExpense   model.Expense
	updatedExpense   model.Expense
	listExpenses     []model.Expense
	listFilters      repository.ListExpenseFilters
	getExpense       model.Expense
	err              error
	deletedExpenseID string
	deletedUserID    string
	deletedBy        string
	getExpenseID     string
	getUserID        string
}

func (s *stubExpenseStore) CreatePersonal(_ context.Context, params repository.CreateExpenseParams) (model.Expense, error) {
	s.createParams = params
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.createdExpense, nil
}

func (s *stubExpenseStore) ListPersonalByUserID(context.Context, string) ([]model.Expense, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.listExpenses, nil
}

func (s *stubExpenseStore) ListPersonal(_ context.Context, filters repository.ListExpenseFilters) ([]model.Expense, error) {
	s.listFilters = filters
	if s.err != nil {
		return nil, s.err
	}
	return s.listExpenses, nil
}

func (s *stubExpenseStore) GetPersonalByID(_ context.Context, expenseID, userID string) (model.Expense, error) {
	s.getExpenseID = expenseID
	s.getUserID = userID
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.getExpense, nil
}

func (s *stubExpenseStore) UpdatePersonal(_ context.Context, params repository.UpdateExpenseParams) (model.Expense, error) {
	s.updateParams = params
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.updatedExpense, nil
}

func (s *stubExpenseStore) DeletePersonal(_ context.Context, expenseID, userID, deletedBy string) error {
	s.deletedExpenseID = expenseID
	s.deletedUserID = userID
	s.deletedBy = deletedBy
	return s.err
}

func (s *stubExpenseStore) RestorePersonal(_ context.Context, expenseID, userID, _ string) (model.Expense, error) {
	s.getExpenseID = expenseID
	s.getUserID = userID
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.getExpense, nil
}
