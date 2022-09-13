package storage

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

// InMemoryRepository is repository that uses memory for storage.
type InMemoryRepository struct {
	storage map[string]models.ShortURL // map that will store urls
	mutex   sync.RWMutex               // read-write mutex that will be used to synchronize access to the storage map
}

// NewInMemoryRepository creates a new InMemoryRepository and returns a pointer to it.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]models.ShortURL),
		mutex:   sync.RWMutex{},
	}
}

// SaveBatch saves multiple urls.
// Checks if the urls are unique and then saving them.
func (repo *InMemoryRepository) SaveBatch(_ context.Context, batch []models.ShortURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {
		_, ok := repo.storage[shortURL.ID]
		if ok {
			return NewNotUniqueURLError(shortURL, nil)
		}
	}

	for _, shortURL := range batch {
		repo.storage[shortURL.ID] = shortURL
	}

	return nil
}

// Save checks if the url is unique and then saving it to the memory.
func (repo *InMemoryRepository) Save(_ context.Context, shortURL models.ShortURL) error {
	repo.mutex.RLock()
	_, ok := repo.storage[shortURL.ID]
	repo.mutex.RUnlock()

	if ok {
		return NewNotUniqueURLError(shortURL, nil)
	}

	repo.mutex.Lock()
	repo.storage[shortURL.ID] = shortURL
	repo.mutex.Unlock()

	return nil
}

// GetByID gets the url by id.
func (repo *InMemoryRepository) GetByID(_ context.Context, id string) (models.ShortURL, error) {
	repo.mutex.RLock()
	url, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if !ok {
		return models.ShortURL{}, errors.New("can't find full url by id")
	}

	return url, nil
}

// GetUsersUrls gets all the urls that were created by the user with the given id.
func (repo *InMemoryRepository) GetUsersUrls(_ context.Context, userID string) ([]models.ShortURL, error) {
	repo.mutex.RLock()
	var URLs []models.ShortURL
	for _, URL := range repo.storage {
		if URL.CreatedByID == userID {
			URLs = append(URLs, URL)
		}
	}
	repo.mutex.RUnlock()
	return URLs, nil
}

// Close clears map.
func (repo *InMemoryRepository) Close(_ context.Context) error {
	repo.storage = make(map[string]models.ShortURL)
	return nil
}

// Check is just a stub.
func (repo *InMemoryRepository) Check(_ context.Context) error {
	return nil
}

// DeleteUrls deletes all given urls.
func (repo *InMemoryRepository) DeleteUrls(_ context.Context, urls []models.ShortURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	now := time.Now()
	for _, urlToDelete := range urls {
		foundURL, ok := repo.storage[urlToDelete.ID]
		if ok && foundURL.CreatedByID == urlToDelete.CreatedByID {
			foundURL.DeletedAt = now
			repo.storage[urlToDelete.ID] = foundURL
		}
	}

	return nil
}
