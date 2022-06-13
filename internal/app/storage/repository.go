package storage

import "github.com/belamov/ypgo-url-shortener/internal/app/config"

type Repository interface {
	Save(url string, id string) error
	GetByID(id string) (string, error)
}

func GetRepo(cfg *config.Config) Repository {
	var repo Repository
	var err error

	if filePath := cfg.FilePath; filePath != "" {
		repo, err = NewFileRepository(filePath)
		if err != nil {
			panic(err)
		}
	} else {
		repo = NewInMemoryRepository()
	}
	return repo
}
