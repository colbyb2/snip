// Package service contains the business logic for Snip.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/colby/snip/internal/model"
	"github.com/colby/snip/internal/repository"
	"github.com/colby/snip/pkg/shortcode"
)

// Common errors returned by the service layer.
var (
	ErrInvalidURL     = errors.New("invalid URL")
	ErrEmptyURL       = errors.New("URL cannot be empty")
	ErrLinkNotFound   = errors.New("link not found")
	ErrCodeGeneration = errors.New("failed to generate unique code after maximum retries")
)

// LinkService handles the business logic for link operations.
type LinkService struct {
	linkRepo    repository.LinkRepository
	clickRepo   repository.ClickRepository
	codeGen     *shortcode.Generator
	baseURL     string
	maxRetries  int
}

// LinkServiceConfig holds configuration for LinkService.
type LinkServiceConfig struct {
	BaseURL       string // e.g., "https://snip.io"
	CodeLength    int    // length of generated short codes
	MaxRetries    int    // max attempts to generate a unique code
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() LinkServiceConfig {
	return LinkServiceConfig{
		BaseURL:    "http://localhost:8080",
		CodeLength: 7,
		MaxRetries: 5,
	}
}

// NewLinkService creates a new LinkService with the given dependencies.
func NewLinkService(
	linkRepo repository.LinkRepository,
	clickRepo repository.ClickRepository,
	config LinkServiceConfig,
) *LinkService {
	return &LinkService{
		linkRepo:   linkRepo,
		clickRepo:  clickRepo,
		codeGen:    shortcode.NewGenerator(config.CodeLength),
		baseURL:    strings.TrimSuffix(config.BaseURL, "/"),
		maxRetries: config.MaxRetries,
	}
}

// CreateLink creates a new shortened URL.
func (s *LinkService) CreateLink(ctx context.Context, originalURL string) (*model.CreateLinkResponse, error) {
	// Validate URL
	if err := s.validateURL(originalURL); err != nil {
		return nil, err
	}

	// Generate unique short code with retry logic
	var link *model.Link
	var err error

	for attempt := 0; attempt < s.maxRetries; attempt++ {
		code, genErr := s.codeGen.Generate()
		if genErr != nil {
			return nil, fmt.Errorf("generating code: %w", genErr)
		}

		link = &model.Link{
			ID:          code, // Using short code as ID for simplicity
			ShortCode:   code,
			OriginalURL: originalURL,
			CreatedAt:   time.Now().UTC(),
			ClickCount:  0,
		}

		err = s.linkRepo.Create(ctx, link)
		if err == nil {
			break // Success!
		}

		if !errors.Is(err, repository.ErrAlreadyExists) {
			return nil, fmt.Errorf("creating link: %w", err)
		}
		// Code collision, retry with new code
	}

	if err != nil {
		return nil, ErrCodeGeneration
	}

	return &model.CreateLinkResponse{
		ShortCode:   link.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, link.ShortCode),
		OriginalURL: link.OriginalURL,
	}, nil
}

// Redirect retrieves the original URL for a short code and records the click.
func (s *LinkService) Redirect(ctx context.Context, shortCode string, metadata ClickMetadata) (string, error) {
	link, err := s.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrLinkNotFound
		}
		return "", fmt.Errorf("fetching link: %w", err)
	}

	// Record click asynchronously to not block redirect
	go s.recordClick(context.Background(), link, metadata)

	return link.OriginalURL, nil
}

// GetStats retrieves statistics for a short code.
func (s *LinkService) GetStats(ctx context.Context, shortCode string) (*model.LinkStats, error) {
	link, err := s.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("fetching link: %w", err)
	}

	return &model.LinkStats{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		ClickCount:  link.ClickCount,
		CreatedAt:   link.CreatedAt,
	}, nil
}

// DeleteLink removes a link by its short code.
func (s *LinkService) DeleteLink(ctx context.Context, shortCode string) error {
	err := s.linkRepo.Delete(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLinkNotFound
		}
		return fmt.Errorf("deleting link: %w", err)
	}
	return nil
}

// ClickMetadata contains information about a redirect request.
type ClickMetadata struct {
	Referrer  string
	UserAgent string
	IPAddress string
}

// recordClick records a click event and increments the counter.
// This runs asynchronously to not block redirects.
func (s *LinkService) recordClick(ctx context.Context, link *model.Link, metadata ClickMetadata) {
	// Increment click count
	_ = s.linkRepo.IncrementClickCount(ctx, link.ShortCode)

	// Record detailed click event
	event := &model.ClickEvent{
		ID:        fmt.Sprintf("%s-%d", link.ShortCode, time.Now().UnixNano()),
		LinkID:    link.ID,
		ClickedAt: time.Now().UTC(),
		Referrer:  metadata.Referrer,
		UserAgent: metadata.UserAgent,
		IPAddress: metadata.IPAddress,
	}

	_ = s.clickRepo.Record(ctx, event)
}

// validateURL checks if the provided URL is valid.
func (s *LinkService) validateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return ErrEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Must have a scheme (http or https)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}

	// Must have a host
	if parsed.Host == "" {
		return ErrInvalidURL
	}

	return nil
}
