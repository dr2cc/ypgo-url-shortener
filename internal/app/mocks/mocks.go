package mocks

import "github.com/stretchr/testify/mock"

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Save(url string, id string) error {
	args := m.Called(url, id)
	return args.Error(0)
}

func (m *MockRepo) GetById(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

type MockGen struct {
	mock.Mock
}

func (m *MockGen) GenerateIdFromString(str string) (string, error) {
	args := m.Called(str)
	return args.String(0), args.Error(1)
}
