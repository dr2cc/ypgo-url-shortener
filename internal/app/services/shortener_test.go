package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestShortener_Expand(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    models.ShortURL
		wantErr bool
	}{
		{
			name: "get full url from id",
			args: args{id: "id"},
			want: models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
			},
			wantErr: false,
		},
		{
			name:    "get full url from missing id",
			args:    args{id: "missing"},
			want:    models.ShortURL{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := new(mocks.MockRepo)
			rm.On("GetByID", "id").Return("url", nil)
			rm.On("GetByID", "missing").Return("", errors.New(""))

			rg := new(mocks.MockRandom)

			service := New(
				rm,
				new(mocks.MockGen),
				rg,
				config.New(),
			)

			got, err := service.Expand(tt.args.id)
			if !tt.wantErr {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShortener_Shorten(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name        string
		args        args
		want        *models.ShortURL
		wantErr     bool
		expectedErr error
	}{
		{
			name: "generate short link from url",
			args: args{url: "url"},
			want: &models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
			},
			wantErr: false,
		},
		{
			name:    "generate short link from empty url",
			args:    args{url: ""},
			want:    &models.ShortURL{},
			wantErr: true,
		},
		{
			name:    "generate short link from url when saving failes",
			args:    args{url: "fail"},
			want:    &models.ShortURL{},
			wantErr: true,
		},
		{
			name:        "it returns correct err when original url is not unique",
			args:        args{url: "fail"},
			want:        &models.ShortURL{},
			wantErr:     true,
			expectedErr: storage.ErrNotUnique,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := new(mocks.MockRepo)
			rm.On("GetByID", "id").Return("url", nil)
			rm.On("GetByID", "missing").Return("", errors.New(""))
			rm.On("Save", models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
			}).Return(nil)
			rm.On("Save", models.ShortURL{
				OriginalURL: "fail",
				ID:          "id",
			}).Return(storage.ErrNotUnique)

			gm := new(mocks.MockGen)
			gm.On("GenerateIDFromString", "url").Return("id", nil)
			gm.On("GenerateIDFromString", "fail").Return("id", nil)
			gm.On("GenerateIDFromString", "").Return("", errors.New(""))

			rg := new(mocks.MockRandom)

			service := New(
				rm,
				gm,
				rg,
				config.New(),
			)

			got, err := service.Shorten(tt.args.url, "")
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			}
			assert.ObjectsAreEqual(tt.want, got)
		})
	}
}

func TestShortener_ShortenBatch(t *testing.T) {
	type args struct {
		batch  []models.ShortURL
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    []responses.ShorteningBatchResult
		wantErr bool
	}{
		{
			name: "short batch urls",
			args: args{
				batch: []models.ShortURL{
					{CorrelationID: "corID", OriginalURL: "origURL"},
					{CorrelationID: "corID2", OriginalURL: "origURL2"},
				},
			},
			want: []responses.ShorteningBatchResult{
				{CorrelationID: "corID", ShortURL: "http://localhost:8080/id"},
				{CorrelationID: "corID2", ShortURL: "http://localhost:8080/id2"},
			},
			wantErr: false,
		},
		{
			name: "short batch urls failed on saving",
			args: args{
				batch: []models.ShortURL{
					{CorrelationID: "corID", OriginalURL: "errorURL"},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := new(mocks.MockRepo)
			rm.On("SaveBatch", tt.args.batch).Return(nil)

			gm := new(mocks.MockGen)
			gm.On("GenerateIDFromString", "origURL").Return("id", nil)
			gm.On("GenerateIDFromString", "origURL2").Return("id2", nil)
			gm.On("GenerateIDFromString", "errorURL").Return("", errors.New(""))

			rg := new(mocks.MockRandom)

			service := New(
				rm,
				gm,
				rg,
				config.New(),
			)

			got, err := service.ShortenBatch(tt.args.batch, tt.args.userID)
			if !tt.wantErr {
				assert.NoError(t, err)
			}
			assert.ObjectsAreEqual(tt.want, got)
		})
	}
}

func TestShortener_GetShortURL(t *testing.T) {
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
			rm := new(mocks.MockRepo)
			gm := new(mocks.MockGen)
			rg := new(mocks.MockRandom)

			service := New(
				rm,
				gm,
				rg,
				cfg,
			)
			model := models.ShortURL{
				ID: tt.fields.ID,
			}

			shortURL := service.FormatShortURL(model.ID)

			assert.Equal(t, tt.want, shortURL)
		})
	}
}
