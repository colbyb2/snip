// Package model defines the core domain types for Snip.
package model

import "time"

// Link represents a shortened URL mapping.
type Link struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int64     `json:"click_count"`
}

// ClickEvent represents a single redirect event for analytics.
type ClickEvent struct {
	ID        string    `json:"id"`
	LinkID    string    `json:"link_id"`
	ClickedAt time.Time `json:"clicked_at"`
	Referrer  string    `json:"referrer,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
}

// CreateLinkRequest represents the input for creating a new short link.
type CreateLinkRequest struct {
	URL string `json:"url"`
}

// CreateLinkResponse represents the output after creating a short link.
type CreateLinkResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// LinkStats represents analytics for a link.
type LinkStats struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}
