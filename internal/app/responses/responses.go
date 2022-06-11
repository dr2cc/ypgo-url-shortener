package responses

import (
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type ShortUrlResponse struct {
	Result string `json:"result"`
}

func NewShortUrlResponse(model models.ShortURL) *ShortUrlResponse {
	return &ShortUrlResponse{Result: model.GetShortUrl()}
}

func (sur *ShortUrlResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
