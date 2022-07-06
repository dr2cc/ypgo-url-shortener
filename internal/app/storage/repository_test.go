package storage

import (
	"os"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestGetRepo(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name string
		args args
		want Repository
	}{
		{
			name: "it returns file repository when file path is present",
			args: args{cfg: &config.Config{
				BaseURL:       "",
				ServerAddress: "",
				FilePath:      "some-path",
			}},
			want: &FileRepository{},
		},
		{
			name: "it returns in memory repository when file path is not present",
			args: args{cfg: &config.Config{
				BaseURL:       "",
				ServerAddress: "",
				FilePath:      "",
			}},
			want: &InMemoryRepository{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsType(t, tt.want, GetRepo(tt.args.cfg))
			if err := os.Remove("some-path"); err != nil {
				return
			}
		})
	}
}
