package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/colby/snip/internal/model"
	"github.com/colby/snip/internal/service"
)

func handleRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	logger.Debug("received request",
		"method", event.RequestContext.HTTP.Method,
		"path", event.RawPath,
	)

	method := event.RequestContext.HTTP.Method
	path := event.RawPath

	switch {
	case method == "GET" && path == "/health":
		return handleHealth()

	case method == "POST" && path == "/api/links":
		return handleCreateLink(ctx, event)

	case method == "GET" && strings.HasPrefix(path, "/api/links/") && strings.HasSuffix(path, "/stats"):
		code := extractCodeFromStatsPath(path)
		return handleGetStats(ctx, code)

	case method == "DELETE" && strings.HasPrefix(path, "/api/links/"):
		code := strings.TrimPrefix(path, "/api/links/")
		return handleDeleteLink(ctx, code)

	case method == "GET" && len(path) > 1:
		code := strings.TrimPrefix(path, "/")
		return handleRedirect(ctx, code, event)

	default:
		return jsonResponse(http.StatusNotFound, map[string]string{"error": "not found"})
	}
}

func extractCodeFromStatsPath(path string) string {
	trimmed := strings.TrimPrefix(path, "/api/links/")
	trimmed = strings.TrimSuffix(trimmed, "/stats")
	return trimmed
}

func handleHealth() (events.APIGatewayV2HTTPResponse, error) {
	return jsonResponse(http.StatusOK, map[string]string{"status": "healthy"})
}

func handleCreateLink(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var req model.CreateLinkRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := linkService.CreateLink(ctx, req.URL)
	if err != nil {
		switch err {
		case service.ErrEmptyURL:
			return jsonResponse(http.StatusBadRequest, map[string]string{"error": "url is required"})
		case service.ErrInvalidURL:
			return jsonResponse(http.StatusBadRequest, map[string]string{"error": "invalid url format"})
		default:
			logger.Error("failed to create link", "error", err)
			return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return jsonResponse(http.StatusCreated, resp)
}

func handleRedirect(ctx context.Context, code string, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	metadata := service.ClickMetadata{
		Referrer:  event.Headers["referer"],
		UserAgent: event.Headers["user-agent"],
		IPAddress: event.RequestContext.HTTP.SourceIP,
	}

	redirectURL, err := linkService.Redirect(ctx, code, metadata)
	if err != nil {
		if err == service.ErrLinkNotFound {
			return jsonResponse(http.StatusNotFound, map[string]string{"error": "link not found"})
		}
		logger.Error("failed to redirect", "code", code, "error", err)
		return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusMovedPermanently,
		Headers: map[string]string{
			"Location": redirectURL,
		},
	}, nil
}

func handleGetStats(ctx context.Context, code string) (events.APIGatewayV2HTTPResponse, error) {
	stats, err := linkService.GetStats(ctx, code)
	if err != nil {
		if err == service.ErrLinkNotFound {
			return jsonResponse(http.StatusNotFound, map[string]string{"error": "link not found"})
		}
		logger.Error("failed to get stats", "code", code, "error", err)
		return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return jsonResponse(http.StatusOK, stats)
}

func handleDeleteLink(ctx context.Context, code string) (events.APIGatewayV2HTTPResponse, error) {
	err := linkService.DeleteLink(ctx, code)
	if err != nil {
		if err == service.ErrLinkNotFound {
			return jsonResponse(http.StatusNotFound, map[string]string{"error": "link not found"})
		}
		logger.Error("failed to delete link", "code", code, "error", err)
		return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func jsonResponse(status int, body any) (events.APIGatewayV2HTTPResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error": "internal server error"}`,
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonBody),
	}, nil
}
