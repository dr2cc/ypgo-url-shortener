package models

type ShortURL struct {
	OriginalURL string `json:"url"`
	ID          string `json:"id"`
	CreatedById string `json:"created_by"`
}
