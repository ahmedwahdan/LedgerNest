package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ahmedwahdan/LedgerNest/backend/internal/config"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/db"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/handler"
	"github.com/ahmedwahdan/LedgerNest/backend/internal/httpx"
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
	authService := service.NewAuthService(userRepository)
	authHandler := handler.NewAuthHandler(authService)

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
