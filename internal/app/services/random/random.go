package random

import (
	"crypto/rand"

	"github.com/google/uuid"
)

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

type UserIDGenerator interface {
	GenerateUserID() string
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) GenerateUserID() string {
	return uuid.NewString()
}
