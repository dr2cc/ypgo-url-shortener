package models

import (
	"fmt"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestShortURL_GetShortURL(t *testing.T) {
	type fields struct {
		OriginalURL string
		ID          string
	}
	cfg := config.New()
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "it correctly generates short url",
			fields: fields{
				ID: "id",
			},
			want: fmt.Sprintf("%s/id", cfg.BaseURL),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ShortURL{
				OriginalURL: tt.fields.OriginalURL,
				ID:          tt.fields.ID,
			}

			shortURL := s.GetShortURL()

			assert.Equal(t, tt.want, shortURL)
		})
	}
}
