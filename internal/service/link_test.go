package service

import (
	"context"
	"strings"
	"testing"

	"github.com/colby/snip/internal/repository"
)

func TestLinkService_CreateLink(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())

	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{
			name:    "valid https URL",
			url:     "https://example.com/path?query=1",
			wantErr: nil,
		},
		{
			name:    "valid http URL",
			url:     "http://example.com",
			wantErr: nil,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: ErrEmptyURL,
		},
		{
			name:    "whitespace only",
			url:     "   ",
			wantErr: ErrEmptyURL,
		},
		{
			name:    "missing scheme",
			url:     "example.com",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "missing host",
			url:     "https://",
			wantErr: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateLink(context.Background(), tt.url)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.ShortCode == "" {
				t.Error("expected non-empty short code")
			}

			if resp.OriginalURL != tt.url {
				t.Errorf("expected original URL %s, got %s", tt.url, resp.OriginalURL)
			}

			if !strings.Contains(resp.ShortURL, resp.ShortCode) {
				t.Errorf("short URL %s should contain short code %s", resp.ShortURL, resp.ShortCode)
			}
		})
	}
}

func TestLinkService_Redirect(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())
	ctx := context.Background()

	// Create a link first
	originalURL := "https://example.com/test"
	resp, err := svc.CreateLink(ctx, originalURL)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Test redirect
	metadata := ClickMetadata{
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	}

	redirectURL, err := svc.Redirect(ctx, resp.ShortCode, metadata)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if redirectURL != originalURL {
		t.Errorf("expected redirect to %s, got %s", originalURL, redirectURL)
	}
}

func TestLinkService_Redirect_NotFound(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())

	_, err := svc.Redirect(context.Background(), "nonexistent", ClickMetadata{})
	if err != ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestLinkService_GetStats(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())
	ctx := context.Background()

	// Create a link
	originalURL := "https://example.com/stats-test"
	resp, err := svc.CreateLink(ctx, originalURL)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Get stats
	stats, err := svc.GetStats(ctx, resp.ShortCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.ShortCode != resp.ShortCode {
		t.Errorf("expected short code %s, got %s", resp.ShortCode, stats.ShortCode)
	}

	if stats.OriginalURL != originalURL {
		t.Errorf("expected original URL %s, got %s", originalURL, stats.OriginalURL)
	}

	if stats.ClickCount != 0 {
		t.Errorf("expected click count 0, got %d", stats.ClickCount)
	}
}

func TestLinkService_GetStats_NotFound(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())

	_, err := svc.GetStats(context.Background(), "nonexistent")
	if err != ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestLinkService_DeleteLink(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())
	ctx := context.Background()

	// Create a link
	resp, err := svc.CreateLink(ctx, "https://example.com/delete-test")
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Delete it
	err = svc.DeleteLink(ctx, resp.ShortCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's gone
	_, err = svc.GetStats(ctx, resp.ShortCode)
	if err != ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound after delete, got %v", err)
	}
}

func TestLinkService_DeleteLink_NotFound(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()
	svc := NewLinkService(linkRepo, clickRepo, DefaultConfig())

	err := svc.DeleteLink(context.Background(), "nonexistent")
	if err != ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestLinkService_CustomBaseURL(t *testing.T) {
	linkRepo := repository.NewMemoryLinkRepository()
	clickRepo := repository.NewMemoryClickRepository()

	config := DefaultConfig()
	config.BaseURL = "https://snip.io/"

	svc := NewLinkService(linkRepo, clickRepo, config)

	resp, err := svc.CreateLink(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(resp.ShortURL, "https://snip.io/") {
		t.Errorf("expected short URL to start with https://snip.io/, got %s", resp.ShortURL)
	}

	// Should not have double slashes
	if strings.Contains(resp.ShortURL, "//"+resp.ShortCode) {
		t.Errorf("short URL has double slashes: %s", resp.ShortURL)
	}
}
