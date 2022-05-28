package generator

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math/big"
)

type HashGenerator struct{}

func (HashGenerator) GenerateIdFromString(str string) (string, error) {
	if str == "" {
		return "", errors.New("empty string to generate id from")
	}

	hash, err := hashUrl(str)
	if err != nil {
		return "", err
	}

	result := toBase62(hash)
	return result, nil
}

func toBase62(id uint32) string {
	var i big.Int
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, id)
	i.SetBytes(b[:])
	return i.Text(62)
}

func hashUrl(url string) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(url))
	if err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}
