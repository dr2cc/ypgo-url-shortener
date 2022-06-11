package requests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenUrlRequest_validation(t *testing.T) {
	tests := []struct {
		name    string
		request *ShortenUrlRequest
	}{
		{
			name:    "it validates that original url is present",
			request: &ShortenUrlRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Bind(&http.Request{})
			assert.Error(t, err)
		})
	}
}

func TestShortenUrlRequest_unmarshalling(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		request *ShortenUrlRequest
	}{
		{
			name:    "it unmarshalls json #1",
			json:    "{\"url\":\"url\"}",
			request: &ShortenUrlRequest{OriginalUrl: "url"},
		},
		{
			name:    "it unmarshalls json #2",
			json:    "{\"url\":\"\"}",
			request: &ShortenUrlRequest{OriginalUrl: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ShortenUrlRequest
			buf := bytes.NewBuffer([]byte(tt.json))
			encoder := json.NewDecoder(buf)
			encoder.Decode(&result)
			assert.ObjectsAreEqual(tt.request, result)
		})
	}
}
