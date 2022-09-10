package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/responses"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
	"github.com/golang/mock/gomock"
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().GetByID(context.Background(), "id").Return(models.ShortURL{OriginalURL: "url", ID: "id"}, nil).AnyTimes()
			mockRepo.EXPECT().GetByID(context.Background(), "missing").Return(models.ShortURL{}, errors.New("")).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)

			mockRandom := mocks.NewMockGenerator(ctrl)

			service := New(mockRepo, mockGen, mockRandom, config.New())

			got, err := service.Expand(context.Background(), tt.args.id)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().Save(context.Background(), models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
			}).Return(nil).AnyTimes()
			mockRepo.EXPECT().Save(context.Background(), models.ShortURL{
				OriginalURL: "fail",
				ID:          "id",
			}).Return(storage.ErrNotUnique).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)
			mockGen.EXPECT().GenerateIDFromString("url").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("fail").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("").Return("", errors.New("")).AnyTimes()

			mockRandom := mocks.NewMockGenerator(ctrl)

			service := New(mockRepo, mockGen, mockRandom, config.New())

			got, err := service.Shorten(context.Background(), tt.args.url, "")
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().SaveBatch(context.Background(), []models.ShortURL{{OriginalURL: "errorURL", CorrelationID: "corID", ID: "id"}}).Return(nil).AnyTimes()
			mockRepo.EXPECT().SaveBatch(context.Background(), []models.ShortURL{
				{CorrelationID: "corID", OriginalURL: "origURL", ID: "id"},
				{CorrelationID: "corID2", OriginalURL: "origURL2", ID: "id2"},
			}).Return(nil).AnyTimes()
			mockRepo.EXPECT().SaveBatch(context.Background(), []models.ShortURL{
				{CorrelationID: "corID", OriginalURL: "errorURL", ID: "id"},
			}).Return(errors.New("")).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)
			mockGen.EXPECT().GenerateIDFromString("origURL").Return("id", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("origURL2").Return("id2", nil).AnyTimes()
			mockGen.EXPECT().GenerateIDFromString("errorURL").Return("", errors.New("")).AnyTimes()

			mockRandom := mocks.NewMockGenerator(ctrl)

			service := New(mockRepo, mockGen, mockRandom, config.New())

			got, err := service.ShortenBatch(context.Background(), tt.args.batch, tt.args.userID)
			if !tt.wantErr {
				assert.NoError(t, err)
			}
			assert.ObjectsAreEqual(tt.want, got)
		})
	}
}

func TestShortener_DeleteUrls(t *testing.T) {
	type args struct {
		urlsIDS []string
		userID  string
	}
	tests := []struct {
		name        string
		args        args
		wantDeleted []models.ShortURL
		wantErr     bool
	}{
		{
			name: "it deletes correct urls",
			args: args{
				urlsIDS: []string{"id1", "id2"},
				userID:  "userID",
			},
			wantDeleted: []models.ShortURL{
				{ID: "id2", CreatedByID: "userID"},
				{ID: "id1", CreatedByID: "userID"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().DeleteUrls(context.Background(), []models.ShortURL{
				{ID: "id2", CreatedByID: "userID"},
				{ID: "id1", CreatedByID: "userID"},
			}).Return(nil).AnyTimes()
			mockRepo.EXPECT().DeleteUrls(context.Background(), []models.ShortURL{
				{ID: "id1", CreatedByID: "userID"},
				{ID: "id2", CreatedByID: "userID"},
			}).Return(nil).AnyTimes()

			mockGen := mocks.NewMockURLGenerator(ctrl)

			mockRandom := mocks.NewMockGenerator(ctrl)

			service := New(mockRepo, mockGen, mockRandom, config.New())

			service.DeleteUrls(context.Background(), tt.args.urlsIDS, tt.args.userID)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockGen := mocks.NewMockURLGenerator(ctrl)
			mockRandom := mocks.NewMockGenerator(ctrl)

			service := New(mockRepo, mockGen, mockRandom, config.New())

			model := models.ShortURL{
				ID: tt.fields.ID,
			}

			shortURL := service.FormatShortURL(model.ID)

			assert.Equal(t, tt.want, shortURL)
		})
	}
}

func BenchmarkShortener(b *testing.B) {
	const urlsToShortenCount = 10000

	urlsToShorten := make([]string, 0)
	for i := 0; i < urlsToShortenCount; i++ {
		urlsToShorten = append(urlsToShorten, randStringBytes(i))
	}

	repo := storage.NewInMemoryRepository()
	gen := generator.HashGenerator{}
	trand := &random.TrulyRandomGenerator{}
	service := New(repo, gen, trand, config.New())

	b.ResetTimer()

	b.Run("shorten", func(b *testing.B) {
		for i := 0; i < urlsToShortenCount; i++ {
			_, _ = service.Shorten(context.Background(), urlsToShorten[i], "")
		}
	})
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))] //nolint:gosec
	}
	return string(b)
}
