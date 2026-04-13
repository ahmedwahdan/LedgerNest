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
	categoryRepository := repository.NewCategoryRepository(pool)
	householdRepository := repository.NewHouseholdRepository(pool)
	budgetCycleRepository := repository.NewBudgetCycleRepository(pool)
	budgetRepository := repository.NewBudgetRepository(pool)
	auditLogRepository := repository.NewAuditLogRepository(pool)
	tokenService := auth.NewTokenService(cfg.JWTSecret)
	auditService := service.NewAuditService(auditLogRepository)
	authService := service.NewAuthService(
		userRepository,
		sessionRepository,
		tokenService,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	expenseService := service.NewExpenseService(expenseRepository, auditService)
	categoryService := service.NewCategoryService(categoryRepository)
	householdService := service.NewHouseholdService(householdRepository, tokenService, cfg.InviteTTL)
	budgetCycleService := service.NewBudgetCycleService(budgetCycleRepository, householdRepository)
	budgetService := service.NewBudgetService(budgetRepository, budgetCycleRepository, householdRepository)
	authHandler := handler.NewAuthHandler(authService)
	expenseHandler := handler.NewExpenseHandler(expenseService, auditService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	householdHandler := handler.NewHouseholdHandler(householdService)
	budgetCycleHandler := handler.NewBudgetCycleHandler(budgetCycleService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
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
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.Handle("GET /auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /expenses", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.ListPersonal)))
	mux.Handle("POST /expenses", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.CreatePersonal)))
	mux.Handle("GET /expenses/{id}", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.GetPersonal)))
	mux.Handle("PUT /expenses/{id}", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.UpdatePersonal)))
	mux.Handle("DELETE /expenses/{id}", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.DeletePersonal)))
	mux.Handle("POST /expenses/{id}/restore", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.RestorePersonal)))
	mux.Handle("GET /expenses/{id}/history", authMiddleware.RequireAuth(http.HandlerFunc(expenseHandler.GetHistory)))
	mux.Handle("GET /categories", authMiddleware.RequireAuth(http.HandlerFunc(categoryHandler.List)))
	mux.Handle("POST /categories", authMiddleware.RequireAuth(http.HandlerFunc(categoryHandler.Create)))
	mux.Handle("PUT /categories/{id}", authMiddleware.RequireAuth(http.HandlerFunc(categoryHandler.Update)))
	mux.Handle("DELETE /categories/{id}", authMiddleware.RequireAuth(http.HandlerFunc(categoryHandler.Delete)))
	mux.Handle("POST /households", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.Create)))
	mux.Handle("GET /households/{id}", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.Get)))
	mux.Handle("PUT /households/{id}", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.Update)))
	mux.Handle("DELETE /households/{id}", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.Delete)))
	mux.Handle("POST /households/{id}/leave", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.Leave)))
	mux.Handle("GET /households/{id}/members", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.ListMembers)))
	mux.Handle("PUT /households/{id}/members/{userId}/role", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.UpdateMemberRole)))
	mux.Handle("DELETE /households/{id}/members/{userId}", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.RemoveMember)))
	mux.Handle("POST /households/{id}/invitations", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.CreateInvitation)))
	mux.Handle("GET /households/{id}/invitations", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.ListInvitations)))
	mux.Handle("DELETE /households/{id}/invitations/{invId}", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.RevokeInvitation)))
	mux.Handle("POST /invitations/accept", authMiddleware.RequireAuth(http.HandlerFunc(householdHandler.AcceptInvitation)))
	mux.Handle("GET /households/{id}/cycle", authMiddleware.RequireAuth(http.HandlerFunc(budgetCycleHandler.GetCycle)))
	mux.Handle("PUT /households/{id}/cycle", authMiddleware.RequireAuth(http.HandlerFunc(budgetCycleHandler.SetCycleConfig)))
	mux.Handle("GET /households/{id}/cycle/snapshots", authMiddleware.RequireAuth(http.HandlerFunc(budgetCycleHandler.ListSnapshots)))
	mux.Handle("GET /budgets", authMiddleware.RequireAuth(http.HandlerFunc(budgetHandler.List)))
	mux.Handle("GET /budgets/health", authMiddleware.RequireAuth(http.HandlerFunc(budgetHandler.GetHealth)))
	mux.Handle("POST /budgets", authMiddleware.RequireAuth(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("PUT /budgets/{id}", authMiddleware.RequireAuth(http.HandlerFunc(budgetHandler.Update)))
	mux.Handle("DELETE /budgets/{id}", authMiddleware.RequireAuth(http.HandlerFunc(budgetHandler.Delete)))

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
