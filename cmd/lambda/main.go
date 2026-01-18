// Package main is the entry point for the Lambda function.
package main

import (
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/colby/snip/internal/service"
)

var linkService *service.LinkService
var logger *slog.Logger

func init() {
	// Setup logger
	logLevel := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	// Get config from environment
	tableName := os.Getenv("DYNAMODB_TABLE")
	baseURL := os.Getenv("BASE_URL")

	if tableName == "" {
		logger.Error("DYNAMODB_TABLE environment variable is required")
		os.Exit(1)
	}

	// Initialize repository
	linkRepo := NewDynamoLinkRepository(tableName)
	clickRepo := NewDynamoClickRepository(tableName)

	// Initialize service
	linkService = service.NewLinkService(linkRepo, clickRepo, service.LinkServiceConfig{
		BaseURL:    baseURL,
		CodeLength: 7,
		MaxRetries: 5,
	})

	logger.Info("lambda initialized", "table", tableName, "base_url", baseURL)
}

func main() {
	lambda.Start(handleRequest)
}
