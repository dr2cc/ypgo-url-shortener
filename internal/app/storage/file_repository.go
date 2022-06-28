package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
)

type FileRepository struct {
	mutex   sync.RWMutex
	file    *os.File
	writer  *bufio.Writer
	scanner *bufio.Scanner
}

func NewFileRepository(filePath string) (*FileRepository, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o777) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		mutex:   sync.RWMutex{},
		file:    file,
		writer:  bufio.NewWriter(file),
		scanner: bufio.NewScanner(file),
	}, nil
}

func (repo *FileRepository) Save(shortURL models.ShortURL) error {
	data, err := json.Marshal(&shortURL)
	if err != nil {
		return err
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, err := repo.writer.Write(data); err != nil {
		return err
	}

	if err := repo.writer.WriteByte('\n'); err != nil {
		return err
	}

	if err := repo.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (repo *FileRepository) GetByID(id string) (models.ShortURL, error) {
	_, err := repo.file.Seek(0, io.SeekStart)
	if err != nil {
		return models.ShortURL{}, err
	}

	var entry models.ShortURL

	for repo.scanner.Scan() {
		line := repo.scanner.Bytes()
		err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry)
		if err != nil {
			return models.ShortURL{}, err
		}
		if entry.ID == id {
			return entry, nil
		}
	}

	return models.ShortURL{}, errors.New("can't find full url by id")
}

func (repo *FileRepository) CloseFile() error {
	return repo.file.Close()
}
