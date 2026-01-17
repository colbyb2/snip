// Package main is the entry point for the Snip API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colby/snip/internal/handler"
	"github.com/colby/snip/internal/repository"
	"github.com/colby/snip/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Configuration (will be from environment variables later)
	cfg := Config{
		Port:       getEnv("PORT", "8080"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		CodeLength: 7,
	}

	// Setup structured logging
	logger := setupLogger(cfg.LogLevel)

	logger.Info("starting snip server",
		"port", cfg.Port,
		"base_url", cfg.BaseURL,
	)

	// Initialize repositories (in-memory for now, will be DynamoDB later)
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()

	// Initialize service
	linkService := service.NewLinkService(linkRepo, clickRepo, service.LinkServiceConfig{
		BaseURL:    cfg.BaseURL,
		CodeLength: cfg.CodeLength,
		MaxRetries: 5,
	})

	// Initialize handlers
	h := handler.New(linkService, logger)

	// Setup HTTP server
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      loggingMiddleware(logger, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		logger.Info("received shutdown signal", "signal", sig)
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

// Config holds server configuration.
type Config struct {
	Port       string
	BaseURL    string
	LogLevel   string
	CodeLength int
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupLogger creates a structured logger with the specified level.
func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	// Use JSON handler for structured logs (better for production/observability)
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

// loggingMiddleware logs HTTP requests.
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"user_agent", r.UserAgent(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
