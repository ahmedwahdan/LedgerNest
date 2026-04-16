package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
)

var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryNameConflict = errors.New("category name already exists in this household")
var ErrSystemCategoryReadOnly = errors.New("system categories cannot be modified")

type categoryStore interface {
	List(ctx context.Context, householdID *string) ([]model.Category, error)
	GetByID(ctx context.Context, categoryID string) (model.Category, error)
	Create(ctx context.Context, params repository.CreateCategoryParams) (model.Category, error)
	Update(ctx context.Context, params repository.UpdateCategoryParams) (model.Category, error)
	Delete(ctx context.Context, categoryID, householdID string) error
}

type CategoryService struct {
	categories categoryStore
	households householdStore
}

func NewCategoryService(categories categoryStore, households householdStore) *CategoryService {
	return &CategoryService{categories: categories, households: households}
}

func (s *CategoryService) List(ctx context.Context, requesterID string, householdID *string) ([]model.Category, error) {
	if householdID != nil {
		if err := s.requireMember(ctx, *householdID, requesterID); err != nil {
			return nil, err
		}
	}

	return s.categories.List(ctx, householdID)
}

func (s *CategoryService) Create(ctx context.Context, requesterID, householdID, name string, parentID, icon, color *string) (model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Category{}, errors.New("name is required")
	}
	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return model.Category{}, err
	}

	return s.categories.Create(ctx, repository.CreateCategoryParams{
		HouseholdID: householdID,
		Name:        name,
		ParentID:    parentID,
		Icon:        icon,
		Color:       color,
	})
}

func (s *CategoryService) Update(ctx context.Context, requesterID, categoryID, householdID, name string, parentID, icon, color *string) (model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Category{}, errors.New("name is required")
	}
	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return model.Category{}, err
	}

	cat, err := s.categories.Update(ctx, repository.UpdateCategoryParams{
		CategoryID:  categoryID,
		HouseholdID: householdID,
		Name:        name,
		ParentID:    parentID,
		Icon:        icon,
		Color:       color,
	})
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return model.Category{}, ErrCategoryNotFound
		}
		if errors.Is(err, repository.ErrCategoryNameConflict) {
			return model.Category{}, ErrCategoryNameConflict
		}
		return model.Category{}, err
	}

	return cat, nil
}

func (s *CategoryService) Delete(ctx context.Context, requesterID, categoryID, householdID string) error {
	if err := s.requireRole(ctx, householdID, requesterID, "owner", "editor"); err != nil {
		return err
	}

	if err := s.categories.Delete(ctx, categoryID, householdID); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}

	return nil
}

func (s *CategoryService) requireMember(ctx context.Context, householdID, userID string) error {
	_, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotMember
		}
		return err
	}

	return nil
}

func (s *CategoryService) requireRole(ctx context.Context, householdID, userID string, roles ...string) error {
	member, err := s.households.GetMembership(ctx, householdID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotMember
		}
		return err
	}

	for _, role := range roles {
		if member.Role == role {
			return nil
		}
	}

	return ErrInsufficientRole
}
