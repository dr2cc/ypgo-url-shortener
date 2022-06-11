package services

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

type Shortener struct {
	repository storage.Repository
	generator  generator.Generator
	config     config.Config
}

func New(repository storage.Repository, generator generator.Generator, config config.Config) *Shortener {
	return &Shortener{
		repository: repository,
		generator:  generator,
		config:     config,
	}
}

func (service *Shortener) Shorten(url string) (models.ShortURL, error) {
	urlID, err := service.generator.GenerateIDFromString(url)
	if err != nil {
		return models.ShortURL{}, err
	}

	err = service.repository.Save(url, urlID)
	if err != nil {
		return models.ShortURL{}, err
	}

	return models.ShortURL{
		OriginalURL: url,
		ID:          urlID,
	}, nil
}

func (service *Shortener) Expand(id string) (string, error) {
	origURL, err := service.repository.GetByID(id)
	if err != nil {
		return "", err
	}
	return origURL, nil
}
