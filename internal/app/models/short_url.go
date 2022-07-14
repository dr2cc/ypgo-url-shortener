package models

import "time"

type ShortURL struct {
	OriginalURL   string    `json:"url"`
	ID            string    `json:"id"`
	CreatedByID   string    `json:"created_by"`
	CorrelationID string    `json:"correlation_id"`
	DeletedAt     time.Time `json:"deleted_at"`
}
