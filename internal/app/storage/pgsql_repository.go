package storage

import (
	"context"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/jackc/pgx/v4"
)

type PgRepository struct {
	Dsn  string
	conn *pgx.Conn
}

func NewPgRepository(dsn string) (*PgRepository, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &PgRepository{
		Dsn:  dsn,
		conn: conn,
	}, nil
}

func (repo *PgRepository) Save(shortURL models.ShortURL) error {
	return nil
}

func (repo *PgRepository) GetByID(id string) (models.ShortURL, error) {
	return models.ShortURL{}, nil
}

func (repo *PgRepository) GetUsersUrls(id string) ([]models.ShortURL, error) {
	var URLs []models.ShortURL
	return URLs, nil
}

func (repo *PgRepository) Close() error {
	return repo.conn.Close(context.Background())
}

func (repo *PgRepository) Check() error {
	return repo.conn.Ping(context.Background())
}
