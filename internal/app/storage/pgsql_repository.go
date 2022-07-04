package storage

import (
	"context"
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/jackc/pgx/v4"
)

const lockNum = int64(9628173550095224)

type PgRepository struct {
	Dsn  string
	conn *pgx.Conn
}

func NewPgRepository(dsn string) (*PgRepository, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	err = runMigrations(conn)
	if err != nil {
		return nil, err
	}
	return &PgRepository{
		Dsn:  dsn,
		conn: conn,
	}, nil
}

func runMigrations(conn *pgx.Conn) error {
	err := acquireAdvisoryLock(context.Background(), conn)
	if err != nil {
		return err
	}
	defer func() {
		unlockErr := releaseAdvisoryLock(context.Background(), conn)
		if err == nil && unlockErr != nil {
			err = unlockErr
		}
	}()

	t, err := conn.Exec(context.Background(), "create table if not exists urls("+
		"created_by varchar(36) not null, "+
		"original_url varchar not null, "+
		"id varchar(12) unique not null"+
		"correlation_id varchar"+
		");")
	fmt.Println(t)
	if err != nil {
		return err
	}

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
	return err
}

func (repo *PgRepository) SaveBatch(batch []models.ShortURL) error {
	_, err := repo.conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"people"},
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
		"select original_url, id, created_by from urls where id=$1",
		id,
	).Scan(&model)
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
		err = rows.Scan(&entry)
		if err != nil {
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

func acquireAdvisoryLock(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, "select pg_advisory_lock($1)", lockNum)
	return err
}

func releaseAdvisoryLock(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, "select pg_advisory_unlock($1)", lockNum)
	return err
}
