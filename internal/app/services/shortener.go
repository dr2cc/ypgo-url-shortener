package services

import (
	"fmt"
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
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

func (service *Shortener) Shorten(url string) (string, error) {
	urlID, err := service.generator.GenerateIdFromString(url)
	if err != nil {
		return "", err
	}

	err = service.repository.Save(url, urlID)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("%s:%s/%s", service.config.Host, service.config.Port, urlID)
	return result, nil
}

func (service *Shortener) Expand(id string) (string, error) {
	origUrl, err := service.repository.GetById(id)
	if err != nil {
		return "", err
	}
	return origUrl, nil
}
