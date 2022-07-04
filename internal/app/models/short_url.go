package models

type ShortURL struct {
	OriginalURL   string `json:"url"`
	ID            string `json:"id"`
	CreatedByID   string `json:"created_by"`
	CorrelationID string `json:"correlation_id"`
}
