package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		c := New()
		assert.Equal(t, ":8080", c.ServerAddress)
		assert.Equal(t, "http://localhost:8080", c.BaseURL)
		assert.Len(t, c.EncryptionKey, 32)
		assert.NotEmpty(t, c.EncryptionKey)
	})
}
