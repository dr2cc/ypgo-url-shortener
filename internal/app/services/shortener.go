package services

import (
	"errors"
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

type shorteningError struct {
	Err      error
	ShortURL models.ShortURL
}

func (err *shorteningError) Error() string {
	return fmt.Sprintf("error while shortening: %v", err.Err)
}

func (err *shorteningError) Unwrap() error {
	return err.Err
}

func NewShorteningError(shortURL models.ShortURL, err error) error {
	return &shorteningError{
		Err:      err,
		ShortURL: shortURL,
	}
}

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
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		return shortURL, NewShorteningError(shortURL, err)
	}
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

func (service *Shortener) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", service.config.BaseURL, urlID)
}

func (service *Shortener) GetUrlsCreatedBy(userID string) ([]models.ShortURL, error) {
	return service.repository.GetUsersUrls(userID)
}

func (service *Shortener) HealthCheck() error {
	return service.repository.Check()
}

func (service *Shortener) ShortenBatch(batch []models.ShortURL, userID string) ([]models.ShortURL, error) {
	for i, URL := range batch {
		urlID, err := service.generator.GenerateIDFromString(URL.OriginalURL)
		if err != nil {
			return nil, err
		}
		batch[i].ID = urlID
		batch[i].CreatedByID = userID
	}

	if err := service.repository.SaveBatch(batch); err != nil {
		return nil, err
	}

	return batch, nil
}
