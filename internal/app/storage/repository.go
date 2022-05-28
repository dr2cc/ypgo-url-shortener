package storage

type Repository interface {
	Save(url string, id string) error
	GetByID(id string) (string, error)
}
