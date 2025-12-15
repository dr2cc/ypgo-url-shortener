package generator // => ./internal/app/services/generator

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math/big"
)

// HashGenerator реализует метод GenerateIDFromString
// интерфейса (generator)generator.URLGenerator
type HashGenerator struct{}

// GenerateIDFromString создает ID (shortURL) из url.
func (HashGenerator) GenerateIDFromString(str string) (string, error) {
	if str == "" {
		return "", errors.New("empty string to generate id from")
	}

	hash, err := hashURL(str)
	if err != nil {
		return "", err
	}

	result := toBase62(hash)
	return result, nil
}

// toBase62 converts a 32-bit integer to a base 62 string.
func toBase62(id uint32) string {
	var i big.Int
	size := 8
	bytes := make([]byte, size)
	binary.LittleEndian.PutUint32(bytes, id)
	i.SetBytes(bytes)
	base := 62
	return i.Text(base)
}

// hashURL takes a string, and returns a 32-bit hash of that string.
func hashURL(url string) (uint32, error) {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(url)); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}
