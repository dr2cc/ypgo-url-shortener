package requests

import (
	"errors"
	"net/http"
)

type ShortenURLRequest struct {
	OriginalURL string `json:"url,omitempty"`
}

func (su *ShortenURLRequest) Bind(r *http.Request) error {
	if su.OriginalURL == "" {
		return errors.New("missing required url field")
	}

	return nil
}
