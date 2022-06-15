package services

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
)

// Decompress распаковывает слайс байт.
func Decompress(data []byte) ([]byte, error) {
	// переменная r будет читать входящие данные и распаковывать их
	r := flate.NewReader(bytes.NewReader(data))
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(r)

	var b bytes.Buffer
	// в переменную b записываются распакованные данные
	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
