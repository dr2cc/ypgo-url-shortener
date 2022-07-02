package storage

import (
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepository_GetByID(t *testing.T) {
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
			name: "get existing id",
			args: args{id: "id"},
			want: models.ShortURL{
				OriginalURL: "some url",
				ID:          "id",
			},
			wantErr: false,
		},
		{
			name:    "get missing id",
			args:    args{id: "missing"},
			want:    models.ShortURL{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryRepository{
				storage: map[string]models.ShortURL{
					"id": {
						OriginalURL: "some url",
						ID:          "id",
					},
				},
			}
			got, err := repo.GetByID(tt.args.id)
			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInMemoryRepository_Save(t *testing.T) {
	type fields struct {
		storage map[string]models.ShortURL
	}
	tests := []struct {
		name        string
		fields      fields
		arg         models.ShortURL
		wantErr     bool
		wantStorage map[string]models.ShortURL
	}{
		{
			name:   "save new url with id",
			fields: fields{storage: map[string]models.ShortURL{}},
			arg: models.ShortURL{
				OriginalURL: "url",
				ID:          "id",
				CreatedByID: "1",
			},
			wantStorage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "save new url with same id",
			fields: fields{
				storage: map[string]models.ShortURL{
					"id": {
						OriginalURL: "some url",
						ID:          "id",
						CreatedByID: "1",
					},
				},
			},
			arg: models.ShortURL{
				OriginalURL: "new url",
				ID:          "id",
				CreatedByID: "1",
			},
			wantStorage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "some url",
					ID:          "id",
					CreatedByID: "1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryRepository{
				storage: tt.fields.storage,
			}
			err := repo.Save(tt.arg)
			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.arg.OriginalURL, repo.storage[tt.arg.ID].OriginalURL)
				assert.Contains(t, repo.storage, tt.arg.ID)
			}
			assert.Equal(t, tt.wantStorage, repo.storage)
		})
	}
}

func TestNewInMemoryRepository(t *testing.T) {
	t.Run("in memory repo init", func(t *testing.T) {
		repo := NewInMemoryRepository()
		assert.Equal(t, &InMemoryRepository{storage: map[string]models.ShortURL{}}, repo)
	})
}
