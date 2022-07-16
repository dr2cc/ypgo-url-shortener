package services

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
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
	Random     random.Generator
	repository storage.Repository
	generator  generator.URLGenerator
	config     *config.Config
}

func New(
	repository storage.Repository,
	generator generator.URLGenerator,
	random random.Generator,
	config *config.Config,
) *Shortener {
	return &Shortener{
		repository: repository,
		generator:  generator,
		config:     config,
		Random:     random,
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

func (service *Shortener) GenerateNewUserID() string {
	return service.Random.GenerateNewUserID()
}

func (service *Shortener) DeleteUrls(ids []string, userID string) {
	// implementing fan-in pattern for learning purposes
	done := make(chan struct{})
	defer close(done)

	workersCount := runtime.NumCPU()
	inputCh := make(chan string)
	modelsToDelete := make([]models.ShortURL, 0, len(ids))

	go func() {
		for _, id := range ids {
			inputCh <- id
		}

		close(inputCh)
	}()

	workerChs := make([]chan models.ShortURL, 0, workersCount)
	for urlID := range inputCh {
		workerCh := make(chan models.ShortURL)
		newWorker(urlID, userID, workerCh)
		workerChs = append(workerChs, workerCh)
	}

	for v := range fanIn(done, workerChs...) {
		modelsToDelete = append(modelsToDelete, v)
	}

	err := service.repository.DeleteUrls(modelsToDelete)
	if err != nil {
		fmt.Printf("couldn't delete urls: %v\n", err)
	}
}

func newWorker(urlID string, userID string, out chan models.ShortURL) {
	go func() {
		defer func() {
			if x := recover(); x != nil {
				newWorker(urlID, userID, out)
				log.Printf("run time panic: %v, %v", x, out)
			}
		}()

		out <- models.ShortURL{ID: urlID, CreatedByID: userID}
		close(out)
	}()
}

func fanIn(done <-chan struct{}, channels ...chan models.ShortURL) chan models.ShortURL {
	var wg sync.WaitGroup
	multiplexedStream := make(chan models.ShortURL)

	multiplex := func(c <-chan models.ShortURL) {
		defer wg.Done()
		for v := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- v:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}
