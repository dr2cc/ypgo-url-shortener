package requests

import (
	"errors"
	"net/http"
)

type ShortenUrlRequest struct {
	OriginalUrl string `json:"url,omitempty"`
}

func (su *ShortenUrlRequest) Bind(r *http.Request) error {
	if su.OriginalUrl == "" {
		return errors.New("missing required url field")
	}

	return nil
}
