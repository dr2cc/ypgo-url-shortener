package storage

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type NotUniqueUrlError struct {
	Err      error
	ShortURL models.ShortURL
}

func (err *NotUniqueUrlError) Error() string {
	return "url or id are already exist"
}

func (err *NotUniqueUrlError) Unwrap() error {
	return err.Err
}

func NewNotUniqueUrlError(shortUrl models.ShortURL, err error) error {
	return &NotUniqueUrlError{
		Err:      err,
		ShortURL: shortUrl,
	}
}

var ErrNotUnique = func() error { return &NotUniqueUrlError{} }()

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
