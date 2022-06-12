package models

import (
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
)

type ShortURL struct {
	OriginalURL string `json:"url"`
	ID          string `json:"-"`
}

func (s ShortURL) GetShortURL() string {
	cfg := config.New()
	return fmt.Sprintf("%s%s", cfg.BaseUrl, s.ID)
}
