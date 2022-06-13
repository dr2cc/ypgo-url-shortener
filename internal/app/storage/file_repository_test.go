package storage

import (
	"os"
	"testing"

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
		want    string
		wantErr bool
	}{
		{
			name:    "get existing id",
			args:    args{id: "id"},
			want:    "url",
			wantErr: false,
		},
		{
			name:    "get missing id",
			args:    args{id: "not existing"},
			want:    "",
			wantErr: true,
		},
	}

	filename := "./test_get_by_id"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		err := repo.CloseFile()
		require.NoError(t, err)
	}(repo)
	defer func(name string) {
		err := os.Remove(name)
		require.NoError(t, err)
	}(filename)

	err = repo.Save("url", "id")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestFileRepository_Save(t *testing.T) {
	type args struct {
		url string
		id  string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantSaved string
	}{
		{
			name: "save new url with id",
			args: args{
				url: "new url",
				id:  "new id",
			},
			wantErr:   false,
			wantSaved: "new url",
		},
		{
			name: "save new url with same id",
			args: args{
				url: "new url",
				id:  "existing id",
			},
			wantSaved: "existing url",
			wantErr:   false,
		},
	}

	filename := "./test_save.json"
	repo, err := NewFileRepository(filename)
	require.NoError(t, err)
	defer func(repo *FileRepository) {
		err := repo.CloseFile()
		require.NoError(t, err)
	}(repo)
	defer func(name string) {
		err := os.Remove(name)
		require.NoError(t, err)
	}(filename)

	err = repo.Save("existing url", "existing id")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(tt.args.url, tt.args.id)
			if tt.wantErr {
				assert.Error(t, err)
				savedURL, err := repo.GetByID(tt.args.id)
				assert.NoError(t, err)
				assert.NotEqual(t, tt.wantSaved, savedURL)
			} else {
				assert.NoError(t, err)
				savedURL, err := repo.GetByID(tt.args.id)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSaved, savedURL)
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
			err := repo.CloseFile()
			assert.NoError(t, err)
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
			err := os.Remove(name)
			require.NoError(t, err)
		}(filename)

		_, err = os.Stat(filename)
		require.NoError(t, err)

		repo, err := NewFileRepository(filename)
		assert.NoError(t, err)

		defer func(repo *FileRepository) {
			err := repo.CloseFile()
			assert.NoError(t, err)
		}(repo)
		assert.NoError(t, err)

		_, err = os.Stat(filename)
		require.NoError(t, err)
	})
}
