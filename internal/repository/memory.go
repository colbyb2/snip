package repository

import (
	"context"
	"sync"

	"github.com/colby/snip/internal/model"
)

// MemoryLinkRepository is an in-memory implementation of LinkRepository.
// Useful for local development and testing.
type MemoryLinkRepository struct {
	mu    sync.RWMutex
	links map[string]*model.Link // keyed by short code
}

// NewMemoryLinkRepository creates a new in-memory link repository.
func NewMemoryLinkRepository() *MemoryLinkRepository {
	return &MemoryLinkRepository{
		links: make(map[string]*model.Link),
	}
}

// Create persists a new link.
func (r *MemoryLinkRepository) Create(ctx context.Context, link *model.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[link.ShortCode]; exists {
		return ErrAlreadyExists
	}

	// Store a copy to avoid external mutations
	stored := *link
	r.links[link.ShortCode] = &stored
	return nil
}

// GetByShortCode retrieves a link by its short code.
func (r *MemoryLinkRepository) GetByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, exists := r.links[shortCode]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to avoid external mutations
	result := *link
	return &result, nil
}

// IncrementClickCount atomically increments the click count.
func (r *MemoryLinkRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, exists := r.links[shortCode]
	if !exists {
		return ErrNotFound
	}

	link.ClickCount++
	return nil
}

// Delete removes a link by its short code.
func (r *MemoryLinkRepository) Delete(ctx context.Context, shortCode string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[shortCode]; !exists {
		return ErrNotFound
	}

	delete(r.links, shortCode)
	return nil
}

// MemoryClickRepository is an in-memory implementation of ClickRepository.
type MemoryClickRepository struct {
	mu     sync.RWMutex
	clicks map[string][]model.ClickEvent // keyed by link ID
}

// NewMemoryClickRepository creates a new in-memory click repository.
func NewMemoryClickRepository() *MemoryClickRepository {
	return &MemoryClickRepository{
		clicks: make(map[string][]model.ClickEvent),
	}
}

// Record persists a new click event.
func (r *MemoryClickRepository) Record(ctx context.Context, event *model.ClickEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clicks[event.LinkID] = append(r.clicks[event.LinkID], *event)
	return nil
}

// GetByLinkID retrieves click events for a link.
func (r *MemoryClickRepository) GetByLinkID(ctx context.Context, linkID string, limit int) ([]model.ClickEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := r.clicks[linkID]
	if len(events) == 0 {
		return []model.ClickEvent{}, nil
	}

	// Return most recent first, up to limit
	if limit <= 0 || limit > len(events) {
		limit = len(events)
	}

	// Copy and return in reverse order (most recent first)
	result := make([]model.ClickEvent, limit)
	for i := 0; i < limit; i++ {
		result[i] = events[len(events)-1-i]
	}

	return result, nil
}
