package models

import (
	"fmt"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestShortUrl_GetShortUrl(t *testing.T) {
	type fields struct {
		OriginalURL string
		Id          string
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
				Id: "id",
			},
			want: fmt.Sprintf("%s:%s/id", cfg.Host, cfg.Port),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ShortURL{
				OriginalURL: tt.fields.OriginalURL,
				Id:          tt.fields.Id,
			}

			shortUrl := s.GetShortUrl()

			assert.Equal(t, tt.want, shortUrl)
		})
	}
}
