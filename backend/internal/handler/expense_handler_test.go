package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

func TestExpenseHandlerCreatePersonal(t *testing.T) {
	t.Parallel()

	handler := NewExpenseHandler(&stubExpenseService{
		createdExpense: model.Expense{ID: "expense-1", Merchant: "Market"},
	})

	request := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString(`{"amount":"12.50","currency":"USD","merchant":"Market","payment_method":"card","date":"2026-04-13"}`))
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.CreatePersonal(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	var response struct {
		Expense model.Expense `json:"expense"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Expense.ID != "expense-1" {
		t.Fatalf("unexpected expense id: %s", response.Expense.ID)
	}
}

func TestExpenseHandlerListPersonal(t *testing.T) {
	t.Parallel()

	stub := &stubExpenseService{
		listExpenses: []model.Expense{{ID: "expense-1"}, {ID: "expense-2"}},
	}
	handler := NewExpenseHandler(stub)

	request := httptest.NewRequest(http.MethodGet, "/expenses?from=2026-04-01&to=2026-04-30&merchant=market&category_id=category-1", http.NoBody)
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.ListPersonal(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if stub.listInput.From != "2026-04-01" {
		t.Fatalf("expected from filter to be forwarded")
	}
}

func TestExpenseHandlerGetPersonal(t *testing.T) {
	t.Parallel()

	handler := NewExpenseHandler(&stubExpenseService{
		getExpense: model.Expense{ID: "expense-1"},
	})

	request := httptest.NewRequest(http.MethodGet, "/expenses/expense-1", http.NoBody)
	request.SetPathValue("id", "expense-1")
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.GetPersonal(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestExpenseHandlerUpdatePersonal(t *testing.T) {
	t.Parallel()

	handler := NewExpenseHandler(&stubExpenseService{
		updatedExpense: model.Expense{ID: "expense-1", Merchant: "Updated"},
	})

	request := httptest.NewRequest(http.MethodPut, "/expenses/expense-1", bytes.NewBufferString(`{"amount":"22.00","currency":"USD","merchant":"Updated","payment_method":"card","date":"2026-04-13"}`))
	request.SetPathValue("id", "expense-1")
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.UpdatePersonal(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestExpenseHandlerDeletePersonal(t *testing.T) {
	t.Parallel()

	handler := NewExpenseHandler(&stubExpenseService{})

	request := httptest.NewRequest(http.MethodDelete, "/expenses/expense-1", http.NoBody)
	request.SetPathValue("id", "expense-1")
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.DeletePersonal(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestExpenseHandlerUpdatePersonalNotFound(t *testing.T) {
	t.Parallel()

	handler := NewExpenseHandler(&stubExpenseService{err: repository.ErrExpenseNotFound})

	request := httptest.NewRequest(http.MethodPut, "/expenses/missing", bytes.NewBufferString(`{"amount":"22.00","currency":"USD","merchant":"Updated","payment_method":"card","date":"2026-04-13"}`))
	request.SetPathValue("id", "missing")
	request = request.WithContext(auth.ContextWithAccessTokenClaims(request.Context(), auth.AccessTokenClaims{UserID: "user-1"}))
	recorder := httptest.NewRecorder()

	handler.UpdatePersonal(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

type stubExpenseService struct {
	createdExpense   model.Expense
	updatedExpense   model.Expense
	listExpenses     []model.Expense
	getExpense       model.Expense
	listInput        service.ListExpensesInput
	err              error
	deletedExpenseID string
	deletedUserID    string
}

func (s *stubExpenseService) CreatePersonal(context.Context, string, service.CreateExpenseInput) (model.Expense, error) {
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.createdExpense, nil
}

func (s *stubExpenseService) ListPersonal(_ context.Context, _ string, input service.ListExpensesInput) ([]model.Expense, error) {
	s.listInput = input
	if s.err != nil {
		return nil, s.err
	}
	return s.listExpenses, nil
}

func (s *stubExpenseService) GetPersonal(context.Context, string, string) (model.Expense, error) {
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.getExpense, nil
}

func (s *stubExpenseService) UpdatePersonal(context.Context, string, string, service.CreateExpenseInput) (model.Expense, error) {
	if s.err != nil {
		return model.Expense{}, s.err
	}
	return s.updatedExpense, nil
}

func (s *stubExpenseService) DeletePersonal(_ context.Context, expenseID, userID string) error {
	s.deletedExpenseID = expenseID
	s.deletedUserID = userID
	return s.err
}
