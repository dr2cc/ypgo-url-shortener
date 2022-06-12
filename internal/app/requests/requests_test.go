package requests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenURLRequest_validation(t *testing.T) {
	tests := []struct {
		name    string
		request *ShortenURLRequest
	}{
		{
			name:    "it validates that original url is present",
			request: &ShortenURLRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Bind(&http.Request{})
			assert.Error(t, err)
		})
	}
}

func TestShortenURLRequest_unmarshalling(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		request *ShortenURLRequest
	}{
		{
			name:    "it unmarshalls json #1",
			json:    "{\"url\":\"url\"}",
			request: &ShortenURLRequest{OriginalURL: "url"},
		},
		{
			name:    "it unmarshalls json #2",
			json:    "{\"url\":\"\"}",
			request: &ShortenURLRequest{OriginalURL: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ShortenURLRequest
			buf := bytes.NewBuffer([]byte(tt.json))
			encoder := json.NewDecoder(buf)
			err := encoder.Decode(&result)
			assert.NoError(t, err)
			assert.ObjectsAreEqual(tt.request, result)
		})
	}
}
