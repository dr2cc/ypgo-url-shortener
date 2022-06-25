package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

type SymmetricCryptographer struct {
	Key []byte
}

func (c *SymmetricCryptographer) Encrypt(src []byte) ([]byte, error) {
	aesblock, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonce := c.Key[len(c.Key)-aesgcm.NonceSize():]

	dst := aesgcm.Seal(nil, nonce, src, nil) // зашифровываем

	return dst, nil
}

func (c *SymmetricCryptographer) Decrypt(src []byte) ([]byte, error) {
	aesblock, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonce := c.Key[len(c.Key)-aesgcm.NonceSize():]

	dst, err := aesgcm.Open(nil, nonce, src, nil) // расшифровываем
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	return dst, nil
}
