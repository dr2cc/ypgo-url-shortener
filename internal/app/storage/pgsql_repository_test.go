package storage

import (
	"context"
	"testing"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func truncate(t time.Time) time.Time {
	return time.Unix(0, t.UnixNano()/1e6*1e6)
}

type PgRepositoryTestSuite struct {
	suite.Suite
	repo *PgRepository
}

func (s *PgRepositoryTestSuite) SetupSuite() {
	repo, err := NewPgRepository("postgres://postgres:postgres@db:5432/praktikum?sslmode=disable", "file://migrations/")
	require.NoError(s.T(), err)
	s.repo = repo
}

func (s *PgRepositoryTestSuite) TearDownTest() {
	_, err := s.repo.conn.Exec(context.Background(), "truncate table urls")
	require.NoError(s.T(), err)
}

func (s *PgRepositoryTestSuite) TearDownSuite() {
	_, err := s.repo.conn.Exec(context.Background(), "truncate table urls")
	require.NoError(s.T(), err)
}

func (s *PgRepositoryTestSuite) TestSave() {
	model := models.ShortURL{
		OriginalURL: "url",
		ID:          "id",
		CreatedByID: "user",
	}
	err := s.repo.Save(model)
	require.NoError(s.T(), err)

	fetched, err := s.repo.GetByID(model.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model, fetched)

	model = models.ShortURL{
		OriginalURL: "url2",
		ID:          "id2",
		CreatedByID: "user",
		DeletedAt:   truncate(time.Now()).UTC(),
	}
	err = s.repo.Save(model)
	require.NoError(s.T(), err)

	fetched, err = s.repo.GetByID(model.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model, fetched)

	model = models.ShortURL{
		OriginalURL:   "url3",
		ID:            "id3",
		CreatedByID:   "user",
		CorrelationID: "cor id",
	}
	err = s.repo.Save(model)
	require.NoError(s.T(), err)

	fetched, err = s.repo.GetByID(model.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model, fetched)

	model = models.ShortURL{
		OriginalURL:   "url4",
		ID:            "id4",
		CreatedByID:   "user",
		CorrelationID: "cor id",
		DeletedAt:     truncate(time.Now()).UTC(),
	}
	err = s.repo.Save(model)
	require.NoError(s.T(), err)

	fetched, err = s.repo.GetByID(model.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), model, fetched)
}

func (s *PgRepositoryTestSuite) TestSaveBatch() {
	m1 := models.ShortURL{
		OriginalURL:   "url",
		ID:            "id",
		CreatedByID:   "user id",
		CorrelationID: "cor id",
		DeletedAt:     truncate(time.Now()).UTC(),
	}
	m2 := models.ShortURL{
		OriginalURL: "url2",
		ID:          "id2",
		CreatedByID: "user2",
		DeletedAt:   truncate(time.Now()).UTC(),
	}
	m3 := models.ShortURL{
		OriginalURL:   "url3",
		ID:            "id3",
		CreatedByID:   "user3",
		CorrelationID: "cor3",
	}
	m4 := models.ShortURL{
		OriginalURL: "url4",
		ID:          "id4",
		CreatedByID: "user4",
	}
	err := s.repo.SaveBatch([]models.ShortURL{m1, m2, m3, m4})
	require.NoError(s.T(), err)

	fetched, err := s.repo.GetByID(m1.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), m1, fetched)

	fetched, err = s.repo.GetByID(m2.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), m2, fetched)

	fetched, err = s.repo.GetByID(m3.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), m3, fetched)

	fetched, err = s.repo.GetByID(m4.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), m4, fetched)
}

func (s *PgRepositoryTestSuite) TestGetUsersUrls() {
	m1 := models.ShortURL{
		OriginalURL:   "url",
		ID:            "id",
		CreatedByID:   "user id",
		CorrelationID: "cor id",
		DeletedAt:     truncate(time.Now()).UTC(),
	}
	m2 := models.ShortURL{
		OriginalURL: "url2",
		ID:          "id2",
		CreatedByID: "user id",
		DeletedAt:   truncate(time.Now()).UTC(),
	}
	m3 := models.ShortURL{
		OriginalURL:   "url3",
		ID:            "id3",
		CreatedByID:   "user3",
		CorrelationID: "cor3",
	}
	m4 := models.ShortURL{
		OriginalURL: "url4",
		ID:          "id4",
		CreatedByID: "user4",
	}
	err := s.repo.SaveBatch([]models.ShortURL{m1, m2, m3, m4})
	require.NoError(s.T(), err)

	fetched, err := s.repo.GetUsersUrls("user id")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), []models.ShortURL{m1, m2}, fetched)
}

func (s *PgRepositoryTestSuite) TestDeleteUrls() {
	m1 := models.ShortURL{
		OriginalURL: "url",
		ID:          "id",
		CreatedByID: "user id",
	}
	m2 := models.ShortURL{
		OriginalURL: "url2",
		ID:          "id2",
		CreatedByID: "user id",
	}
	m3 := models.ShortURL{
		OriginalURL: "url3",
		ID:          "id3",
		CreatedByID: "user3",
	}
	m4 := models.ShortURL{
		OriginalURL: "url4",
		ID:          "id4",
		CreatedByID: "user4",
	}
	err := s.repo.SaveBatch([]models.ShortURL{m1, m2, m3, m4})
	require.NoError(s.T(), err)

	err = s.repo.DeleteUrls([]models.ShortURL{m1, m2})
	assert.NoError(s.T(), err)

	fetched, err := s.repo.GetByID(m1.ID)
	assert.NoError(s.T(), err)
	assert.False(s.T(), fetched.DeletedAt.IsZero())

	fetched, err = s.repo.GetByID(m2.ID)
	assert.NoError(s.T(), err)
	assert.False(s.T(), fetched.DeletedAt.IsZero())

	modelWithWrongUserID := models.ShortURL{
		OriginalURL: m3.OriginalURL,
		ID:          m3.ID,
		CreatedByID: "another user",
	}
	err = s.repo.DeleteUrls([]models.ShortURL{modelWithWrongUserID})
	assert.NoError(s.T(), err)

	fetched, err = s.repo.GetByID(modelWithWrongUserID.ID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), fetched.DeletedAt.IsZero())

	err = s.repo.DeleteUrls([]models.ShortURL{})
	assert.NoError(s.T(), err)

	fetched, err = s.repo.GetByID(m4.ID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), fetched.DeletedAt.IsZero())
}

func (s *PgRepositoryTestSuite) TestGetById() {
	fetched, err := s.repo.GetByID("not existing")
	assert.Error(s.T(), err)
	assert.Equal(s.T(), models.ShortURL{}, fetched)
}

func TestPgRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PgRepositoryTestSuite))
}
