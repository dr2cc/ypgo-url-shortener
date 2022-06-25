package crypto

import (
	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSymmetricCryptographer_encryption(t *testing.T) {
	t.Run("it encrypts and decrypts message", func(t *testing.T) {
		c := &SymmetricCryptographer{
			Key: config.New().EncryptionKey,
		}

		msg := "some message"
		encrypted, err := c.Encrypt([]byte(msg))
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := c.Decrypt(encrypted)
		require.NoError(t, err)

		assert.Equal(t, []byte(msg), decrypted)
		assert.Equal(t, msg, string(decrypted))
	})
}
