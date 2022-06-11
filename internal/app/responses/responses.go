package responses

import (
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type ShortURLResponse struct {
	Result string `json:"result"`
}

func NewShortURLResponse(model models.ShortURL) *ShortURLResponse {
	return &ShortURLResponse{Result: model.GetShortURL()}
}

func (sur *ShortURLResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
