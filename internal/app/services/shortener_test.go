package services

import (
	"errors"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestShortener_Expand(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "get full url from id",
			args:    args{id: "id"},
			want:    "url",
			wantErr: false,
		},
		{
			name:    "get full url from missing id",
			args:    args{id: "missing"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := new(mocks.MockRepo)
			rm.On("GetByID", "id").Return("url", nil)
			rm.On("GetByID", "missing").Return("", errors.New(""))

			service := New(
				rm,
				new(mocks.MockGen),
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
		name    string
		args    args
		want    *models.ShortURL
		wantErr bool
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := new(mocks.MockRepo)
			rm.On("GetByID", "id").Return("url", nil)
			rm.On("GetByID", "missing").Return("", errors.New(""))
			rm.On("Save", "url", "id").Return(nil)
			rm.On("Save", "fail", "id").Return(errors.New(""))

			gm := new(mocks.MockGen)
			gm.On("GenerateIDFromString", "url").Return("id", nil)
			gm.On("GenerateIDFromString", "fail").Return("id", nil)
			gm.On("GenerateIDFromString", "").Return("", errors.New(""))

			service := New(
				rm,
				gm,
				config.New(),
			)

			got, err := service.Shorten(tt.args.url)
			if !tt.wantErr {
				assert.NoError(t, err)
			}
			assert.ObjectsAreEqual(tt.want, got)
		})
	}
}
