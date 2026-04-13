package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/auth"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/config"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/db"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/handler"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/middleware"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/repository"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()

	pool, err := db.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	userRepository := repository.NewUserRepository(pool)
	sessionRepository := repository.NewSessionRepository(pool)
	expenseRepository := repository.NewExpenseRepository(pool)
	tokenService := auth.NewTokenService(cfg.JWTSecret)
	authService := service.NewAuthService(
		userRepository,
		sessionRepository,
		tokenService,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	expenseService := service.NewExpenseService(expenseRepository)
	authHandler := handler.NewAuthHandler(authService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		healthCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := pool.Ping(healthCtx); err != nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.Handle("GET /auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /expenses", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.ListPersonal)))
	mux.Handle("POST /expenses", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.CreatePersonal)))
	mux.Handle("GET /expenses/{id}", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.GetPersonal)))
	mux.Handle("PUT /expenses/{id}", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.UpdatePersonal)))
	mux.Handle("DELETE /expenses/{id}", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.DeletePersonal)))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("server starting", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
