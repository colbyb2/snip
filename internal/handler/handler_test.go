package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/colby/snip/internal/model"
	"github.com/colby/snip/internal/repository"
	"github.com/colby/snip/internal/service"
)

func setupTestHandler() (*Handler, *http.ServeMux) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	linkService := service.NewLinkService(linkRepo, clickRepo, service.DefaultConfig())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	h := New(linkService, logger)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	return h, mux
}

func TestHandler_CreateLink(t *testing.T) {
	_, mux := setupTestHandler()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid URL",
			body:       `{"url": "https://example.com"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty body",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid URL",
			body:       `{"url": "not-a-url"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}

			if tt.wantStatus == http.StatusCreated {
				var resp model.CreateLinkResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.ShortCode == "" {
					t.Error("expected non-empty short code")
				}
			}
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	_, mux := setupTestHandler()

	// First create a link
	createReq := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(`{"url": "https://example.com/target"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	var createResp model.CreateLinkResponse
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Test redirect
	redirectReq := httptest.NewRequest(http.MethodGet, "/"+createResp.ShortCode, nil)
	redirectRec := httptest.NewRecorder()
	mux.ServeHTTP(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, redirectRec.Code)
	}

	location := redirectRec.Header().Get("Location")
	if location != "https://example.com/target" {
		t.Errorf("expected location https://example.com/target, got %s", location)
	}
}

func TestHandler_Redirect_NotFound(t *testing.T) {
	_, mux := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestHandler_GetStats(t *testing.T) {
	_, mux := setupTestHandler()

	// First create a link
	createReq := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(`{"url": "https://example.com/stats"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	var createResp model.CreateLinkResponse
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Get stats
	statsReq := httptest.NewRequest(http.MethodGet, "/api/links/"+createResp.ShortCode+"/stats", nil)
	statsRec := httptest.NewRecorder()
	mux.ServeHTTP(statsRec, statsReq)

	if statsRec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, statsRec.Code, statsRec.Body.String())
	}

	var stats model.LinkStats
	if err := json.NewDecoder(statsRec.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats response: %v", err)
	}

	if stats.ShortCode != createResp.ShortCode {
		t.Errorf("expected short code %s, got %s", createResp.ShortCode, stats.ShortCode)
	}
}

func TestHandler_DeleteLink(t *testing.T) {
	_, mux := setupTestHandler()

	// First create a link
	createReq := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(`{"url": "https://example.com/delete"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	var createResp model.CreateLinkResponse
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Delete the link
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/links/"+createResp.ShortCode, nil)
	deleteRec := httptest.NewRecorder()
	mux.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
	}

	// Verify it's gone
	statsReq := httptest.NewRequest(http.MethodGet, "/api/links/"+createResp.ShortCode+"/stats", nil)
	statsRec := httptest.NewRecorder()
	mux.ServeHTTP(statsRec, statsReq)

	if statsRec.Code != http.StatusNotFound {
		t.Errorf("expected status %d after delete, got %d", http.StatusNotFound, statsRec.Code)
	}
}

func TestHandler_HealthCheck(t *testing.T) {
	_, mux := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp["status"])
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4"},
			remoteAddr: "5.6.7.8:12345",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4, 5.6.7.8"},
			remoteAddr: "9.10.11.12:12345",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "1.2.3.4"},
			remoteAddr: "5.6.7.8:12345",
			want:       "1.2.3.4",
		},
		{
			name:       "fallback to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "1.2.3.4:12345",
			want:       "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("expected %s, got %s", tt.want, got)
			}
		})
	}
}
