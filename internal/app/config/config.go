package config

import (
	"crypto/aes"
	"flag"
	"os"

	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
)

type Config struct {
	BaseURL       string
	ServerAddress string
	FilePath      string
	EncryptionKey []byte
	DatabaseDSN   string
}

func New() *Config {
	size := 2 * aes.BlockSize //nolint:gomnd
	randomKey, err := random.GenerateRandom(size)
	if err != nil {
		randomKey = make([]byte, size)
	}
	return &Config{
		BaseURL:       "http://localhost:8080",
		ServerAddress: ":8080",
		FilePath:      "",
		EncryptionKey: randomKey,
		DatabaseDSN:   "",
	}
}

func (c *Config) Init() {
	flag.StringVar(&c.ServerAddress, "a", getEnv("SERVER_ADDRESS", ":8080"), "host to listen on")
	flag.StringVar(&c.BaseURL, "b", getEnv("BASE_URL", "http://localhost:8080"), "base url")
	flag.StringVar(&c.FilePath, "f", getEnv("FILE_STORAGE_PATH", ""), "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", getEnv("DATABASE_DSN", ""), "database dsn for connecting to postgres")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
