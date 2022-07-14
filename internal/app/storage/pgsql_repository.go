package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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

	if err = runMigrations(dsn); err != nil {
		return nil, err
	}
	return &PgRepository{
		Dsn:  dsn,
		conn: conn,
	}, nil
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file://internal/app/storage/migrations/", dsn)
	if err != nil {
		return err
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("Nothing to migrate")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println("Migrated successfully")
	return nil
}

func (repo *PgRepository) Save(shortURL models.ShortURL) error {
	_, err := repo.conn.Exec(
		context.Background(),
		"insert into urls (original_url, id, created_by) values ($1, $2, $3)",
		shortURL.OriginalURL,
		shortURL.ID,
		shortURL.CreatedByID,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return NewNotUniqueURLError(shortURL, err)
		}
	}
	return err
}

func (repo *PgRepository) SaveBatch(batch []models.ShortURL) error {
	_, err := repo.conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"urls"},
		[]string{"original_url", "id", "created_by", "correlation_id"},
		pgx.CopyFromSlice(len(batch), func(i int) ([]interface{}, error) {
			return []interface{}{batch[i].OriginalURL, batch[i].ID, batch[i].CreatedByID, batch[i].CorrelationID}, nil
		}),
	)
	return err
}

func (repo *PgRepository) GetByID(id string) (models.ShortURL, error) {
	var model models.ShortURL
	err := repo.conn.QueryRow(
		context.Background(),
		"select original_url, id, created_by, correlation_id from urls where id=$1",
		id,
	).Scan(&model.OriginalURL, &model.ID, &model.CreatedByID, &model.CorrelationID)
	return model, err
}

func (repo *PgRepository) GetUsersUrls(userID string) ([]models.ShortURL, error) {
	var URLs []models.ShortURL

	rows, err := repo.conn.Query(
		context.Background(),
		"select original_url, id, created_by from urls where created_by=$1",
		userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var entry models.ShortURL
		if err = rows.Scan(&entry); err != nil {
			return nil, err
		}
		URLs = append(URLs, entry)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return URLs, nil
}

func (repo *PgRepository) Close() error {
	return repo.conn.Close(context.Background())
}

func (repo *PgRepository) Check() error {
	return repo.conn.Ping(context.Background())
}

func (repo *PgRepository) DeleteUrls(urls []models.ShortURL) error {
	if len(urls) == 0 {
		return nil
	}
	deletedAt := time.Now()
	urlIds := make([]string, len(urls))
	userId := urls[0].CreatedByID

	for i, url := range urls {
		urlIds[i] = url.ID
	}

	_, err := repo.conn.Exec(
		context.Background(),
		"update urls set deleted_at = $1 where created_by = $2 and id = any($3)",
		deletedAt,
		userId,
		urlIds,
	)

	return err
}
