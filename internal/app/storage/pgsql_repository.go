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
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type PgRepository struct {
	conn *pgx.Conn // connection to the database
	Dsn  string    // data source name for the Postgres database. It's a string that contains the host, port, username, password, and database name
}

// NewPgRepository creates a new Postgres connection, runs the migrations, and returns a new PgRepository.
func NewPgRepository(dsn string, migrationsPath string) (*PgRepository, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err = runMigrations(dsn, migrationsPath); err != nil {
		return nil, err
	}
	return &PgRepository{
		Dsn:  dsn,
		conn: conn,
	}, nil
}

// runMigrations runs migrations that hasn't been run.
func runMigrations(dsn string, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, dsn)
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

// Save inserting a new row into the urls table.
func (repo *PgRepository) Save(ctx context.Context, shortURL models.ShortURL) error {
	_, err := repo.conn.Exec(
		ctx,
		"insert into urls (original_url, id, created_by, correlation_id, deleted_at) values ($1, $2, $3, $4, $5)",
		shortURL.OriginalURL,
		shortURL.ID,
		shortURL.CreatedByID,
		shortURL.CorrelationID,
		shortURL.DeletedAt,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return NewNotUniqueURLError(shortURL, err)
		}
	}
	return err
}

// SaveBatch is a batch insert operation.
func (repo *PgRepository) SaveBatch(ctx context.Context, batch []models.ShortURL) error {
	_, err := repo.conn.CopyFrom(
		ctx,
		pgx.Identifier{"urls"},
		[]string{"original_url", "id", "created_by", "correlation_id", "deleted_at"},
		pgx.CopyFromSlice(len(batch), func(i int) ([]interface{}, error) {
			return []interface{}{batch[i].OriginalURL, batch[i].ID, batch[i].CreatedByID, batch[i].CorrelationID, batch[i].DeletedAt}, nil
		}),
	)
	return err
}

// GetByID gets url by id.
func (repo *PgRepository) GetByID(ctx context.Context, id string) (models.ShortURL, error) {
	var model models.ShortURL
	var deletedAt pgtype.Timestamp
	var correlationID pgtype.Text
	err := repo.conn.QueryRow(
		ctx,
		"select original_url, id, created_by, correlation_id, deleted_at from urls where id=$1",
		id,
	).Scan(&model.OriginalURL, &model.ID, &model.CreatedByID, &correlationID, &deletedAt)
	model.DeletedAt = deletedAt.Time
	model.CorrelationID = correlationID.String
	return model, err
}

// GetUsersUrls returns all the urls created by a user.
func (repo *PgRepository) GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error) {
	var URLs []models.ShortURL

	rows, err := repo.conn.Query(
		ctx,
		"select original_url, id, created_by, correlation_id, deleted_at from urls where created_by=$1",
		userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		model := models.ShortURL{}
		var deletedAt pgtype.Timestamp
		var correlationID pgtype.Text
		if err = rows.Scan(&model.OriginalURL, &model.ID, &model.CreatedByID, &correlationID, &deletedAt); err != nil {
			return nil, err
		}
		model.DeletedAt = deletedAt.Time
		model.CorrelationID = correlationID.String
		URLs = append(URLs, model)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return URLs, nil
}

// Close closes the connection to the database.
func (repo *PgRepository) Close(ctx context.Context) error {
	return repo.conn.Close(ctx)
}

// Check checks if the database is up and running.
func (repo *PgRepository) Check(ctx context.Context) error {
	return repo.conn.Ping(ctx)
}

// DeleteUrls deletes all given urls.
func (repo *PgRepository) DeleteUrls(ctx context.Context, urls []models.ShortURL) error {
	if len(urls) == 0 {
		return nil
	}
	deletedAt := time.Now()
	urlsToDelete := make(map[string][]string)

	for _, url := range urls {
		urlsToDelete[url.CreatedByID] = append(urlsToDelete[url.CreatedByID], url.ID)
	}

	for userID, urlIDs := range urlsToDelete {
		if _, err := repo.conn.Exec(
			ctx,
			"update urls set deleted_at = $1 where created_by = $2 and id = any($3)",
			deletedAt,
			userID,
			urlIDs,
		); err != nil {
			return err
		}
	}

	return nil
}
