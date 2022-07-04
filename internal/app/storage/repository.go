package storage

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type Repository interface {
	Save(shortURL models.ShortURL) error
	GetByID(id string) (models.ShortURL, error)
	GetUsersUrls(id string) ([]models.ShortURL, error)
	Close() error
	Check() error
	SaveBatch(batch []models.ShortURL) error
}

func GetRepo(cfg *config.Config) Repository {
	if cfg.DatabaseDSN != "" {
		repo, err := NewPgRepository(cfg.DatabaseDSN)
		if err != nil {
			panic(err)
		}
		return repo
	}
	if cfg.FilePath != "" {
		repo, err := NewFileRepository(cfg.FilePath)
		if err != nil {
			panic(err)
		}
		return repo
	}

	return NewInMemoryRepository()
}
