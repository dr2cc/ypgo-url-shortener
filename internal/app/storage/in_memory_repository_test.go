package storage

import (
	"context"
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
			got, err := repo.GetByID(context.Background(), tt.args.id)
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
		fields      fields
		wantStorage map[string]models.ShortURL
		arg         models.ShortURL
		name        string
		wantErr     bool
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
			err := repo.Save(context.Background(), tt.arg)
			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.arg.OriginalURL, repo.storage[tt.arg.ID].OriginalURL)
				assert.Contains(t, repo.storage, tt.arg.ID)
			}
			assert.Equal(t, tt.wantStorage, repo.storage)
		})
	}
}

func TestInMemoryRepository_SaveBatch(t *testing.T) {
	type fields struct {
		storage map[string]models.ShortURL
	}
	tests := []struct {
		fields      fields
		wantStorage map[string]models.ShortURL
		name        string
		arg         []models.ShortURL
		wantErr     bool
	}{
		{
			name:   "save new url with id",
			fields: fields{storage: map[string]models.ShortURL{}},
			arg: []models.ShortURL{
				{
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "1",
				},
				{
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "2",
				},
			},
			wantStorage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "1",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "2",
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
			arg: []models.ShortURL{
				{
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "1",
				},
				{
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "2",
				},
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
			err := repo.SaveBatch(context.Background(), tt.arg)
			if !tt.wantErr {
				require.NoError(t, err)
			}
			assert.ObjectsAreEqual(tt.wantStorage, repo.storage)
		})
	}
}

func TestNewInMemoryRepository(t *testing.T) {
	t.Run("in memory repo init", func(t *testing.T) {
		repo := NewInMemoryRepository()
		assert.Equal(t, &InMemoryRepository{storage: map[string]models.ShortURL{}}, repo)
	})
}

func TestInMemoryRepository_GetUsersUrls(t *testing.T) {
	type fields struct {
		storage map[string]models.ShortURL
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []models.ShortURL
	}{
		{
			name: "it returns urls of user",
			fields: struct{ storage map[string]models.ShortURL }{storage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
			}},
			args: args{id: "user"},
			want: []models.ShortURL{{
				OriginalURL: "url",
				ID:          "id",
				CreatedByID: "user",
			}},
		},
		{
			name: "it returns null of non existing user",
			fields: struct{ storage map[string]models.ShortURL }{storage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
			}},
			args: args{id: "non existing user"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryRepository{
				storage: tt.fields.storage,
			}
			URLs, _ := repo.GetUsersUrls(context.Background(), tt.args.id)
			assert.Equal(t, tt.want, URLs)
			assert.Equal(t, len(tt.want), len(URLs))
		})
	}
}

func TestInMemoryRepository_DeleteUrls(t *testing.T) {
	type fields struct {
		storage map[string]models.ShortURL
	}
	type args struct {
		urls []models.ShortURL
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantNotDeleted []string
		wantDeleted    []string
	}{
		{
			name: "it deletes urls correctly",
			fields: struct{ storage map[string]models.ShortURL }{storage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
				"id3": {
					OriginalURL: "url3",
					ID:          "id3",
					CreatedByID: "user3",
				},
			}},
			args: args{urls: []models.ShortURL{
				{
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				{
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
			}},
			wantNotDeleted: []string{"id3"},
			wantDeleted:    []string{"id", "id2"},
		},
		{
			name: "it deletes urls correctly when empty array is provided",
			fields: struct{ storage map[string]models.ShortURL }{storage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
				"id3": {
					OriginalURL: "url3",
					ID:          "id3",
					CreatedByID: "user3",
				},
			}},
			args:           args{urls: []models.ShortURL{}},
			wantNotDeleted: []string{"id", "id2", "id3"},
			wantDeleted:    []string{},
		},
		{
			name: "it deletes urls correctly when provided more urls than exists",
			fields: struct{ storage map[string]models.ShortURL }{storage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
				"id3": {
					OriginalURL: "url3",
					ID:          "id3",
					CreatedByID: "user3",
				},
			}},
			args: args{urls: []models.ShortURL{
				{
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				{
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
				{
					OriginalURL: "url3",
					ID:          "id3",
					CreatedByID: "user3",
				},
				{
					OriginalURL: "url4",
					ID:          "id4",
					CreatedByID: "user4",
				},
			}},
			wantNotDeleted: []string{},
			wantDeleted:    []string{"id", "id2", "id3"},
		},
		{
			name: "it doesn't delete urls of wrong user",
			fields: struct{ storage map[string]models.ShortURL }{storage: map[string]models.ShortURL{
				"id": {
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user",
				},
				"id2": {
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user2",
				},
			}},
			args: args{urls: []models.ShortURL{
				{
					OriginalURL: "url",
					ID:          "id",
					CreatedByID: "user2",
				},
				{
					OriginalURL: "url2",
					ID:          "id2",
					CreatedByID: "user",
				},
			}},

			wantNotDeleted: []string{"id", "id2"},
			wantDeleted:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryRepository{
				storage: tt.fields.storage,
			}
			err := repo.DeleteUrls(context.Background(), tt.args.urls)
			assert.NoError(t, err)

			for _, id := range tt.wantDeleted {
				assert.False(t, repo.storage[id].DeletedAt.IsZero())
			}
			for _, id := range tt.wantNotDeleted {
				assert.True(t, repo.storage[id].DeletedAt.IsZero())
			}
		})
	}
}
