package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		c := New()
		assert.Equal(t, "8080", c.Port)
		assert.Equal(t, "http://localhost", c.Host)
	})
}
