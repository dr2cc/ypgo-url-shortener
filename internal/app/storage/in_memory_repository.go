package storage

import "errors"

type InMemoryRepository struct {
	storage storage
}

type storage map[string]string

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(storage),
	}
}

func (repo *InMemoryRepository) Save(url string, id string) error {
	if _, ok := repo.storage[id]; ok {
		return errors.New("not unique id")
	}
	repo.storage[id] = url
	return nil
}

func (repo *InMemoryRepository) GetByID(id string) (string, error) {
	fu, ok := repo.storage[id]
	if !ok {
		return "", errors.New("can't find full url by id")
	}
	return fu, nil
}
