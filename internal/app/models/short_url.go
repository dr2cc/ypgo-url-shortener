// Package models contains business models description.
package models

import "time"

// ShortURL is main entity for system.
type ShortURL struct {
	DeletedAt     time.Time `json:"deleted_at"`     // is used to mark a record as deleted
	OriginalURL   string    `json:"url"`            // original URL that was shortened
	ID            string    `json:"id"`             // unique ID of the short URL.
	CreatedByID   string    `json:"created_by"`     // ID of the user who created the short URL
	CorrelationID string    `json:"correlation_id"` // CorrelationID is used for matching original and shorten urls in shorten batch operation
}
