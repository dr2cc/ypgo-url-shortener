package config

import (
	"crypto/rand"
	"flag"
	"os"
)

type Config struct {
	BaseURL       string
	ServerAddress string
	FilePath      string
	EncryptionKey []byte
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func New() *Config {
	randomKey, err := generateRandom(16)
	if err != nil {
		randomKey = make([]byte, 16)
	}
	return &Config{
		BaseURL:       "http://localhost:8080",
		ServerAddress: ":8080",
		FilePath:      "",
		EncryptionKey: randomKey,
	}
}

func (c *Config) Init() {
	flag.StringVar(&c.ServerAddress, "a", getEnv("SERVER_ADDRESS", ":8080"), "host to listen on")
	flag.StringVar(&c.BaseURL, "b", getEnv("BASE_URL", "http://localhost:8080"), "base url")
	flag.StringVar(&c.FilePath, "f", getEnv("FILE_STORAGE_PATH", ""), "file storage path")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
