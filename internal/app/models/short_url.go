package models

import (
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
)

type ShortURL struct {
	OriginalURL string `json:"url"`
	ID          string `json:"id"`
	CreatedById string `json:"-"`
	Cfg         *config.Config
}

func (s ShortURL) GetShortURL() string {
	return fmt.Sprintf("%s/%s", s.Cfg.BaseURL, s.ID)
}
