package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
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
