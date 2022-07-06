package generator

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math/big"
)

type HashGenerator struct{}

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

func toBase62(id uint32) string {
	var i big.Int
	size := 8
	bytes := make([]byte, size)
	binary.LittleEndian.PutUint32(bytes, id)
	i.SetBytes(bytes)
	base := 62
	return i.Text(base)
}

func hashURL(url string) (uint32, error) {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(url)); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}
