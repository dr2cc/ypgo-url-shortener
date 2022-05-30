package storage

import (
	"errors"
	"sync"
)

type InMemoryRepository struct {
	storage map[string]string
	mutex   sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]string),
		mutex:   sync.RWMutex{},
	}
}

func (repo *InMemoryRepository) Save(url string, id string) error {
	repo.mutex.RLock()
	_, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if ok {
		return errors.New("not unique id")
	}

	repo.mutex.Lock()
	repo.storage[id] = url
	repo.mutex.Unlock()

	return nil
}

func (repo *InMemoryRepository) GetByID(id string) (string, error) {
	repo.mutex.RLock()
	url, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if !ok {
		return "", errors.New("can't find full url by id")
	}

	return url, nil
}
