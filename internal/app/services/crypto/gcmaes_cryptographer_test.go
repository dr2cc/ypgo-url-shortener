package crypto

import (
	"crypto/aes"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymmetricCryptographer_encryption(t *testing.T) {
	t.Run("it encrypts and decrypts message", func(t *testing.T) {
		c := &GCMAESCryptographer{
			Key:    make([]byte, 2*aes.BlockSize),
			Random: &random.TrulyRandomGenerator{},
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
