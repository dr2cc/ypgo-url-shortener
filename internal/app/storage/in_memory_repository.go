package storage

import (
	"errors"
	"sync"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type InMemoryRepository struct {
	storage map[string]models.ShortURL
	mutex   sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]models.ShortURL),
		mutex:   sync.RWMutex{},
	}
}

func (repo *InMemoryRepository) SaveBatch(batch []models.ShortURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {
		_, ok := repo.storage[shortURL.ID]
		if ok {
			return errors.New("not unique id " + shortURL.ID)
		}
	}

	for _, shortURL := range batch {
		repo.storage[shortURL.ID] = shortURL
	}

	return nil
}

func (repo *InMemoryRepository) Save(shortURL models.ShortURL) error {
	repo.mutex.RLock()
	_, ok := repo.storage[shortURL.ID]
	repo.mutex.RUnlock()

	if ok {
		return errors.New("not unique id")
	}

	repo.mutex.Lock()
	repo.storage[shortURL.ID] = shortURL
	repo.mutex.Unlock()

	return nil
}

func (repo *InMemoryRepository) GetByID(id string) (models.ShortURL, error) {
	repo.mutex.RLock()
	url, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if !ok {
		return models.ShortURL{}, errors.New("can't find full url by id")
	}

	return url, nil
}

func (repo *InMemoryRepository) GetUsersUrls(id string) ([]models.ShortURL, error) {
	repo.mutex.RLock()
	var URLs []models.ShortURL
	for _, URL := range repo.storage {
		if URL.CreatedByID == id {
			URLs = append(URLs, URL)
		}
	}
	repo.mutex.RUnlock()
	return URLs, nil
}

func (repo *InMemoryRepository) Close() error {
	repo.storage = make(map[string]models.ShortURL)
	return nil
}

func (repo *InMemoryRepository) Check() error {
	return nil
}
