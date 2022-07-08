package random

import (
	"crypto/rand"

	"github.com/google/uuid"
)

type Generator interface {
	GenerateRandomBytes(size int) ([]byte, error)
	GenerateNewUserID() string
}

type TrulyRandomGenerator struct{}

func (g *TrulyRandomGenerator) GenerateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

func (g *TrulyRandomGenerator) GenerateNewUserID() string {
	return uuid.NewString()
}
