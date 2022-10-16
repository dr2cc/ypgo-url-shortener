package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		c, err := New()
		require.NoError(t, err)
		assert.Equal(t, ":8080", c.ServerAddress)
		assert.Equal(t, "http://localhost:8080", c.BaseURL)
		assert.Len(t, c.EncryptionKey, 32)
		assert.NotEmpty(t, c.EncryptionKey)
	})
}
