package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
)

type GCMAESCryptographer struct {
	Key []byte
}

func (c *GCMAESCryptographer) Encrypt(plaintext []byte) ([]byte, error) {
	aesblock, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonce, err := random.GenerateRandom(aesgcm.NonceSize())
	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (c *GCMAESCryptographer) Decrypt(ciphertext []byte) ([]byte, error) {
	aesblock, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext is too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return aesgcm.Open(nil, nonce, ciphertext, nil) // расшифровываем
}
