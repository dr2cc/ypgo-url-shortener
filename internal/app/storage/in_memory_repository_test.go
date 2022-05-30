package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepository_GetById(t *testing.T) {
	type fields struct {
		storage map[string]string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get existing id",
			fields: fields{storage: map[string]string{
				"id": "some url",
			}},
			args:    args{id: "id"},
			want:    "some url",
			wantErr: false,
		},
		{
			name: "get missing url",
			fields: fields{storage: map[string]string{
				"id": "some url",
			}},
			args:    args{id: "missing"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryRepository{
				storage: tt.fields.storage,
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
		storage map[string]string
	}
	type args struct {
		url string
		id  string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantStorage map[string]string
	}{
		{
			name:   "save new url with id",
			fields: fields{storage: map[string]string{}},
			args:   args{id: "id", url: "url"},
			wantStorage: map[string]string{
				"id": "url",
			},
			wantErr: false,
		},
		{
			name: "save new url with id",
			fields: fields{
				storage: map[string]string{
					"id": "some url",
				},
			},
			args: args{id: "id", url: "new url"},
			wantStorage: map[string]string{
				"id": "some url",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemoryRepository{
				storage: tt.fields.storage,
			}
			err := repo.Save(tt.args.url, tt.args.id)
			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, tt.args.url, repo.storage[tt.args.id])
				assert.Contains(t, repo.storage, tt.args.id)
			}
			assert.Equal(t, tt.wantStorage, repo.storage)
		})
	}
}

func TestNewInMemoryRepository(t *testing.T) {
	t.Run("in memory repo init", func(t *testing.T) {
		repo := NewInMemoryRepository()
		assert.Equal(t, &InMemoryRepository{storage: map[string]string{}}, repo)
	})
}
