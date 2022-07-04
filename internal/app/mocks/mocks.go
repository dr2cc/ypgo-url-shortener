package mocks

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) SaveBatch(batch []models.ShortURL) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *MockRepo) GetUsersUrls(id string) ([]models.ShortURL, error) {
	args := m.Called(id)
	if args.String(0) == "" && args.String(1) == "" {
		return []models.ShortURL{}, nil
	}
	return []models.ShortURL{{OriginalURL: args.String(0), ID: args.String(1), CreatedByID: id}}, nil
}

func (m *MockRepo) Save(shortURL models.ShortURL) error {
	args := m.Called(shortURL)
	return args.Error(0)
}

func (m *MockRepo) GetByID(id string) (models.ShortURL, error) {
	args := m.Called(id)
	return models.ShortURL{OriginalURL: args.String(0), ID: id}, args.Error(1)
}

func (m *MockRepo) Close() error {
	return nil
}

func (m *MockRepo) Check() error {
	return nil
}

type MockGen struct {
	mock.Mock
}

func (m *MockGen) GenerateIDFromString(str string) (string, error) {
	args := m.Called(str)
	return args.String(0), args.Error(1)
}

type MockUserIDGenerator struct {
	mock.Mock
}

func (m *MockUserIDGenerator) GenerateUserID() string {
	args := m.Called()
	return args.String(0)
}
