package storage

type Repository interface {
	Save(url string, id string) error
	GetById(id string) (string, error)
}
