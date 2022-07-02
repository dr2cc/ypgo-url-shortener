package services

import (
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

type Shortener struct {
	repository storage.Repository
	generator  generator.Generator
	config     *config.Config
}

func New(repository storage.Repository, generator generator.Generator, config *config.Config) *Shortener {
	return &Shortener{
		repository: repository,
		generator:  generator,
		config:     config,
	}
}

func (service *Shortener) Shorten(url string, userID string) (models.ShortURL, error) {
	urlID, err := service.generator.GenerateIDFromString(url)
	if err != nil {
		return models.ShortURL{}, err
	}

	shortURL := models.ShortURL{
		OriginalURL: url,
		ID:          urlID,
		CreatedByID: userID,
	}

	err = service.repository.Save(shortURL)
	if err != nil {
		return models.ShortURL{}, err
	}

	return shortURL, nil
}

func (service *Shortener) Expand(id string) (models.ShortURL, error) {
	origURL, err := service.repository.GetByID(id)
	if err != nil {
		return models.ShortURL{}, err
	}
	return origURL, nil
}

func (service *Shortener) FormatShortURL(model models.ShortURL) string {
	return fmt.Sprintf("%s/%s", service.config.BaseURL, model.ID)
}

func (service *Shortener) GetUrlsCreatedBy(userID string) ([]models.ShortURL, error) {
	return service.repository.GetUsersUrls(userID)
}
