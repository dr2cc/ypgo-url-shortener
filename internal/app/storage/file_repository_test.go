package storage

import (
	"context"
	"os"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_GetByID(t *testing.T) {
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
				OriginalURL: "url",
				ID:          "id",
			},
			wantErr: false,
		},
		{
			name:    "get missing id",
			args:    args{id: "not existing"},
			want:    models.ShortURL{},
			wantErr: true,
		},
	}

	filename := "./test_get_by_id"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		errClose := repo.Close(context.Background())
		require.NoError(t, errClose)
	}(repo)
	defer func(name string) {
		errRemove := os.Remove(name)
		require.NoError(t, errRemove)
	}(filename)

	err = repo.Save(context.Background(), models.ShortURL{
		OriginalURL: "url",
		ID:          "id",
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGet := repo.GetByID(context.Background(), tt.args.id)
			if !tt.wantErr {
				require.NoError(t, errGet)
			} else {
				assert.Error(t, errGet)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileRepository_Save(t *testing.T) {
	tests := []struct {
		arg       models.ShortURL
		wantSaved models.ShortURL
		name      string
		wantErr   bool
	}{
		{
			name: "save new url with id",
			arg: models.ShortURL{
				OriginalURL: "new url",
				ID:          "new id",
			},
			wantErr: false,
			wantSaved: models.ShortURL{
				OriginalURL: "new url",
				ID:          "new id",
			},
		},
		{
			name: "save new url with same id",
			arg: models.ShortURL{
				OriginalURL: "new url",
				ID:          "existing id",
			},
			wantSaved: models.ShortURL{
				OriginalURL: "existing url",
				ID:          "existing id",
			},
			wantErr: true,
		},
	}

	filename := "./test_save.json"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		errClose := repo.Close(context.Background())
		require.NoError(t, errClose)
	}(repo)
	defer func(name string) {
		errRemove := os.Remove(name)
		require.NoError(t, errRemove)
	}(filename)

	err = repo.Save(context.Background(), models.ShortURL{
		OriginalURL: "existing url",
		ID:          "existing id",
		CreatedByID: "",
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errSave := repo.Save(context.Background(), tt.arg)
			if tt.wantErr {
				assert.Error(t, errSave)
			} else {
				assert.NoError(t, errSave)
			}
			savedURL, errGet := repo.GetByID(context.Background(), tt.arg.ID)
			assert.NoError(t, errGet)
			assert.Equal(t, tt.wantSaved, savedURL)
		})
	}
}

func TestFileRepository_SaveBatch(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		arg       []models.ShortURL
		wantSaved []models.ShortURL
		wantErr   bool
	}{
		{
			name: "save new urls",
			arg: []models.ShortURL{
				{
					OriginalURL: "new url",
					ID:          "new id",
					CreatedByID: "user",
				},
				{
					OriginalURL: "new url2",
					ID:          "new id2",
					CreatedByID: "user",
				},
			},
			wantErr: false,
			wantSaved: []models.ShortURL{
				{
					OriginalURL: "new url",
					ID:          "new id",
					CreatedByID: "user",
				},
				{
					OriginalURL: "new url2",
					ID:          "new id2",
					CreatedByID: "user",
				},
			},
			userID: "user",
		},
		{
			name: "save new urls with same id",
			arg: []models.ShortURL{
				{
					OriginalURL: "new url",
					ID:          "new id",
					CreatedByID: "user2",
				},
				{
					OriginalURL: "new url2",
					ID:          "new id2",
					CreatedByID: "user2",
				},
			},
			wantSaved: nil,
			wantErr:   true,
			userID:    "use2r",
		},
	}

	filename := "./test_save_batch.json"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		errClose := repo.Close(context.Background())
		require.NoError(t, errClose)
	}(repo)
	defer func(name string) {
		errRemove := os.Remove(name)
		require.NoError(t, errRemove)
	}(filename)

	err = repo.Save(context.Background(), models.ShortURL{
		OriginalURL: "existing url",
		ID:          "existing id",
		CreatedByID: "user existed",
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errSave := repo.SaveBatch(context.Background(), tt.arg)
			if tt.wantErr {
				assert.Error(t, errSave)
			} else {
				assert.NoError(t, errSave)
			}
			savedURLs, errGet := repo.GetUsersUrls(context.Background(), tt.userID)
			assert.NoError(t, errGet)
			assert.Equal(t, tt.wantSaved, savedURLs)
		})
	}
}

func TestFileRepository_DeleteUrls(t *testing.T) {
	type fields struct {
		storage []models.ShortURL
	}
	type args struct {
		urls []models.ShortURL
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantSaved   []models.ShortURL
		wantDeleted []models.ShortURL
	}{
		{
			name: "it deletes urls correctly",
			fields: struct{ storage []models.ShortURL }{storage: []models.ShortURL{
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
			wantSaved: []models.ShortURL{
				{
					OriginalURL: "url3",
					ID:          "id3",
					CreatedByID: "user3",
				},
			},
			wantDeleted: []models.ShortURL{
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
			},
		},
		{
			name: "it deletes urls correctly when empty array is provided",
			fields: struct{ storage []models.ShortURL }{storage: []models.ShortURL{
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
			}},
			args: args{urls: []models.ShortURL{}},
			wantSaved: []models.ShortURL{
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
			},
			wantDeleted: []models.ShortURL{},
		},
		{
			name: "it deletes urls correctly when provided more urls than exists",
			fields: struct{ storage []models.ShortURL }{storage: []models.ShortURL{
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
			wantSaved: []models.ShortURL{},
			wantDeleted: []models.ShortURL{
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
			},
		},
		{
			name: "it doesn't delete urls of wrong user",
			fields: struct{ storage []models.ShortURL }{storage: []models.ShortURL{
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
			wantSaved: []models.ShortURL{
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
			},
			wantDeleted: []models.ShortURL{},
		},
	}

	filename := "./test_delete.json"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		errClose := repo.Close(context.Background())
		require.NoError(t, errClose)
	}(repo)
	defer func(name string) {
		errRemove := os.Remove(name)
		require.NoError(t, errRemove)
	}(filename)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = repo.file.Truncate(0)
			require.NoError(t, err)
			_, err = repo.file.Seek(0, 0)
			require.NoError(t, err)
			err = repo.SaveBatch(context.Background(), tt.fields.storage)
			require.NoError(t, err)

			err = repo.DeleteUrls(context.Background(), tt.args.urls)
			assert.NoError(t, err)

			for _, url := range tt.wantDeleted {
				deleted, errGet := repo.GetByID(context.Background(), url.ID)
				assert.NoError(t, errGet)
				assert.False(t, deleted.DeletedAt.IsZero())
			}

			for _, url := range tt.wantSaved {
				foundURL, errGet := repo.GetByID(context.Background(), url.ID)
				assert.NoError(t, errGet)
				assert.True(t, foundURL.DeletedAt.IsZero())
			}
		})
	}
}

func TestNewFileRepository(t *testing.T) {
	t.Run("it creates file if t doesnt exists", func(t *testing.T) {
		filename := "./not_existing_file.json"
		defer func(name string) {
			err := os.Remove(name)
			require.NoError(t, err)
		}(filename)

		_, err := os.Stat(filename)
		require.ErrorIs(t, err, os.ErrNotExist)

		repo, err := NewFileRepository(filename)
		defer func(repo *FileRepository) {
			errClose := repo.Close(context.Background())
			assert.NoError(t, errClose)
		}(repo)

		require.NoError(t, err)
		_, err = os.Stat(filename)
		require.NoError(t, err)
	})
	t.Run("it initializes from existing file", func(t *testing.T) {
		filename := "./existing_file.json"
		_, err := os.Create(filename)
		require.NoError(t, err)
		defer func(name string) {
			errRemove := os.Remove(name)
			require.NoError(t, errRemove)
		}(filename)

		_, err = os.Stat(filename)
		require.NoError(t, err)

		repo, err := NewFileRepository(filename)
		assert.NoError(t, err)

		defer func(repo *FileRepository) {
			errClose := repo.Close(context.Background())
			assert.NoError(t, errClose)
		}(repo)
		assert.NoError(t, err)

		_, err = os.Stat(filename)
		require.NoError(t, err)
	})
}

func TestFileRepository_GetUsersUrls(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    []models.ShortURL
		wantErr bool
	}{
		{
			name: "get existing user id",
			args: args{userID: "user id"},
			want: []models.ShortURL{{
				OriginalURL: "url",
				ID:          "id",
				CreatedByID: "user id",
			}},
		},
		{
			name: "get missing user id",
			args: args{userID: "not existing"},
			want: nil,
		},
	}
	filename := "./test_get_by_user_id"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		errClose := repo.Close(context.Background())
		require.NoError(t, errClose)
	}(repo)
	defer func(name string) {
		errRemove := os.Remove(name)
		require.NoError(t, errRemove)
	}(filename)

	err = repo.Save(context.Background(), models.ShortURL{
		OriginalURL: "url",
		ID:          "id",
		CreatedByID: "user id",
	})
	require.NoError(t, err)

	err = repo.Save(context.Background(), models.ShortURL{
		OriginalURL: "url2",
		ID:          "id2",
		CreatedByID: "user2 id",
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGet := repo.GetUsersUrls(context.Background(), tt.args.userID)
			require.NoError(t, errGet)
			assert.Equal(t, tt.want, got)
		})
	}
}
