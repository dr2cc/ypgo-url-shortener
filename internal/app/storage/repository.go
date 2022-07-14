package storage

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type NotUniqueURLError struct {
	Err      error
	ShortURL models.ShortURL
}

func (err *NotUniqueURLError) Error() string {
	return "url or id are already exist"
}

func (err *NotUniqueURLError) Unwrap() error {
	return err.Err
}

func NewNotUniqueURLError(shortURL models.ShortURL, err error) error {
	return &NotUniqueURLError{
		Err:      err,
		ShortURL: shortURL,
	}
}

var ErrNotUnique = func() error { return &NotUniqueURLError{} }()

type Repository interface {
	Save(shortURL models.ShortURL) error
	GetByID(id string) (models.ShortURL, error)
	GetUsersUrls(id string) ([]models.ShortURL, error)
	Close() error
	Check() error
	SaveBatch(batch []models.ShortURL) error
	DeleteUrls(urls []models.ShortURL) error
}

func GetRepo(cfg *config.Config) Repository {
	if cfg.DatabaseDSN != "" {
		repo, err := NewPgRepository(cfg.DatabaseDSN, cfg.MigrationsPath)
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
