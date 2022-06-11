package models

import (
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
)

type ShortURL struct {
	OriginalURL string `json:"url,omitempty"`
	Id          string `json:"result,omitempty"`
}

func (s ShortURL) GetShortUrl() string {
	cfg := config.New()
	return fmt.Sprintf("%s:%s/%s", cfg.Host, cfg.Port, s.Id)
}
