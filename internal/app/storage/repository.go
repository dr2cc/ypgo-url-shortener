package storage

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type Repository interface {
	Save(shortURL models.ShortURL) error
	GetByID(id string) (models.ShortURL, error)
}

func GetRepo(cfg *config.Config) Repository {
	var repo Repository
	var err error

	if filePath := cfg.FilePath; filePath != "" {
		repo, err = NewFileRepository(filePath)
		if err != nil {
			panic(err)
		}
	} else {
		repo = NewInMemoryRepository()
	}
	return repo
}
