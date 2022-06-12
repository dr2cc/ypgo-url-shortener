package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		c := New()
		assert.Equal(t, "http://localhost:8080", c.ServerAddress)
		assert.Equal(t, "http://localhost:8080/", c.BaseURL)
	})
}
