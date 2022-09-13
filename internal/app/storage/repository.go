// Package storage contains logic for saving and retrieving data.
package storage

import (
	"context"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

// NotUniqueURLError is error occurred when saving url is already exists.
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

// Repository saves and retrieves data from storage.
type Repository interface {
	Save(ctx context.Context, shortURL models.ShortURL) error
	GetByID(ctx context.Context, id string) (models.ShortURL, error)
	GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error)
	Close(_ context.Context) error
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []models.ShortURL) error
	DeleteUrls(ctx context.Context, urls []models.ShortURL) error
}

// GetRepo is fabric that returns repository implementation based on cfg.
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
