package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terrafertil/backend-go/internal/adapter/ai"
	"terrafertil/backend-go/internal/api"
	"terrafertil/backend-go/internal/service"
	"terrafertil/backend-go/internal/store"
)

func main() {
	port := getenv("PORT", "8080")
	aiServiceURL := getenv("AI_SERVICE_URL", "http://localhost:8001")
	databasePath := getenv("DATABASE_PATH", "./terrafertil.db")
	allowedOrigin := getenv("ALLOWED_ORIGIN", "*")

	// 1. Persistencia SQLite con pool afinado
	db, err := store.OpenSQLite(databasePath)
	if err != nil {
		api.Logger.Error("could not open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	waitlistStore, err := store.NewWaitlistStore(db)
	if err != nil {
		api.Logger.Error("could not initialize waitlist store", "error", err)
		os.Exit(1)
	}

	// 2. Adapters & Services
	projectService := service.NewProjectService()
	aiClient := ai.NewResilientAIClient(aiServiceURL, 3, 10*time.Second)

	// 3. Routing & Middlewares
	mux := http.NewServeMux()
	server := api.NewServer(aiClient, waitlistStore, projectService)
	server.RegisterRoutes(mux)

	handler := api.RequestIDMiddleware(
		api.ObservabilityMiddleware(
			api.SecurityHeadersMiddleware(allowedOrigin)(mux),
		),
	)

	// 4. Servidor HTTP de producción con timeouts estrictos
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 5. Arranque y Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		api.Logger.Info("backend-go listening",
			"port", port,
			"ai_service", aiServiceURL,
			"database", databasePath,
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			api.Logger.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()

	<-stopChan
	api.Logger.Info("shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		api.Logger.Error("forced shutdown error", "error", err)
	}
	api.Logger.Info("backend-go stopped cleanly")
}

func getenv(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
