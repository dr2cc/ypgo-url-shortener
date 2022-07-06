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
	mutex  sync.RWMutex
	file   *os.File
	writer *bufio.Writer
}

func NewFileRepository(filePath string) (*FileRepository, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o777) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		mutex:  sync.RWMutex{},
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (repo *FileRepository) SaveBatch(batch []models.ShortURL) error {
	for _, shortURL := range batch {
		_, err := repo.GetByID(shortURL.ID)
		if err == nil {
			return NewNotUniqueURLError(shortURL, nil)
		}
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {

		data, err := json.Marshal(shortURL)
		if err != nil {
			return err
		}

		if _, err := repo.writer.Write(data); err != nil {
			return err
		}

		if err := repo.writer.WriteByte('\n'); err != nil {
			return err
		}

	}
	if err := repo.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (repo *FileRepository) Save(shortURL models.ShortURL) error {
	_, err := repo.GetByID(shortURL.ID)
	if err == nil {
		return NewNotUniqueURLError(shortURL, nil)
	}

	data, err := json.Marshal(shortURL)
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
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return models.ShortURL{}, err
	}

	var entry models.ShortURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return models.ShortURL{}, err
		}
		if entry.ID == id {
			return entry, nil
		}
	}

	return models.ShortURL{}, errors.New("can't find full url by id")
}

func (repo *FileRepository) GetUsersUrls(id string) ([]models.ShortURL, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var entry models.ShortURL
	var URLs []models.ShortURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return nil, err
		}
		if entry.CreatedByID == id {
			URLs = append(URLs, entry)
		}
	}

	return URLs, nil
}

func (repo *FileRepository) Close() error {
	return repo.file.Close()
}

func (repo *FileRepository) Check() error {
	_, err := repo.file.Stat()
	return err
}
