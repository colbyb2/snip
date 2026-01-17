// Package handler contains HTTP handlers for the Snip API.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/colby/snip/internal/model"
	"github.com/colby/snip/internal/service"
)

// Handler holds the HTTP handlers and their dependencies.
type Handler struct {
	linkService *service.LinkService
	logger      *slog.Logger
}

// New creates a new Handler with the given dependencies.
func New(linkService *service.LinkService, logger *slog.Logger) *Handler {
	return &Handler{
		linkService: linkService,
		logger:      logger,
	}
}

// RegisterRoutes registers all HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/links", h.CreateLink)
	mux.HandleFunc("GET /api/links/{code}/stats", h.GetStats)
	mux.HandleFunc("DELETE /api/links/{code}", h.DeleteLink)
	mux.HandleFunc("GET /{code}", h.Redirect)
	mux.HandleFunc("GET /health", h.HealthCheck)
}

// CreateLink handles POST /api/links
func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.linkService.CreateLink(r.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyURL):
			h.writeError(w, http.StatusBadRequest, "url is required")
		case errors.Is(err, service.ErrInvalidURL):
			h.writeError(w, http.StatusBadRequest, "invalid url format")
		default:
			h.logger.Error("failed to create link", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// Redirect handles GET /{code}
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		h.writeError(w, http.StatusBadRequest, "short code is required")
		return
	}

	metadata := service.ClickMetadata{
		Referrer:  r.Header.Get("Referer"),
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: getClientIP(r),
	}

	redirectURL, err := h.linkService.Redirect(r.Context(), code, metadata)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			h.writeError(w, http.StatusNotFound, "link not found")
			return
		}
		h.logger.Error("failed to redirect", "code", code, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

// GetStats handles GET /api/links/{code}/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		h.writeError(w, http.StatusBadRequest, "short code is required")
		return
	}

	stats, err := h.linkService.GetStats(r.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			h.writeError(w, http.StatusNotFound, "link not found")
			return
		}
		h.logger.Error("failed to get stats", "code", code, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// DeleteLink handles DELETE /api/links/{code}
func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		h.writeError(w, http.StatusBadRequest, "short code is required")
		return
	}

	err := h.linkService.DeleteLink(r.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			h.writeError(w, http.StatusNotFound, "link not found")
			return
		}
		h.logger.Error("failed to delete link", "code", code, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// writeJSON writes a JSON response with the given status code.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

// writeError writes a JSON error response.
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{
		"error": message,
	})
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (common for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
