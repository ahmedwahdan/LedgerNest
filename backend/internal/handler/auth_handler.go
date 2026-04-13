package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/model"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

type authService interface {
	Register(ctx context.Context, email, password, displayName string) (model.User, error)
	Login(ctx context.Context, email, password string) (service.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (service.LoginResult, error)
	CurrentUser(ctx context.Context, userID string) (model.User, error)
}

type AuthHandler struct {
	auth authService
}

func NewAuthHandler(auth authService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateRegisterRequest(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, repository.ErrUserEmailExists) {
			httpx.WriteError(w, http.StatusConflict, "email already in use")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateLoginRequest(req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to authenticate user")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	result, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to refresh session")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.AccessTokenClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	user, err := h.auth.CurrentUser(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid access token")
			return
		}

		httpx.WriteError(w, http.StatusInternalServerError, "failed to load authenticated user")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
}

func validateRegisterRequest(req registerRequest) error {
	switch {
	case strings.TrimSpace(req.Email) == "":
		return errors.New("email is required")
	case !strings.Contains(req.Email, "@"):
		return errors.New("email is invalid")
	case len(req.Password) < 8:
		return errors.New("password must be at least 8 characters")
	case strings.TrimSpace(req.DisplayName) == "":
		return errors.New("display_name is required")
	default:
		return nil
	}
}

func validateLoginRequest(req loginRequest) error {
	switch {
	case strings.TrimSpace(req.Email) == "":
		return errors.New("email is required")
	case !strings.Contains(req.Email, "@"):
		return errors.New("email is invalid")
	case strings.TrimSpace(req.Password) == "":
		return errors.New("password is required")
	default:
		return nil
	}
}
