// Package repository defines interfaces for data persistence.
package repository

import (
	"context"
	"errors"

	"github.com/colby/snip/internal/model"
)

// Common errors returned by repository implementations.
var (
	ErrNotFound      = errors.New("link not found")
	ErrAlreadyExists = errors.New("short code already exists")
)

// LinkRepository defines the interface for link persistence operations.
// This abstraction allows us to swap implementations (in-memory, DynamoDB, PostgreSQL)
// without changing the service layer.
type LinkRepository interface {
	// Create persists a new link. Returns ErrAlreadyExists if the short code is taken.
	Create(ctx context.Context, link *model.Link) error

	// GetByShortCode retrieves a link by its short code. Returns ErrNotFound if not found.
	GetByShortCode(ctx context.Context, shortCode string) (*model.Link, error)

	// IncrementClickCount atomically increments the click count for a link.
	IncrementClickCount(ctx context.Context, shortCode string) error

	// Delete removes a link by its short code.
	Delete(ctx context.Context, shortCode string) error
}

// ClickRepository defines the interface for click event persistence.
type ClickRepository interface {
	// Record persists a new click event.
	Record(ctx context.Context, event *model.ClickEvent) error

	// GetByLinkID retrieves all click events for a given link.
	GetByLinkID(ctx context.Context, linkID string, limit int) ([]model.ClickEvent, error)
}
