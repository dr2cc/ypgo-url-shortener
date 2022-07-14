package config

import (
	"crypto/aes"
	"flag"
	"os"

	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
)

const KeySize = 2 * aes.BlockSize //nolint:gomnd

type Config struct {
	BaseURL        string
	ServerAddress  string
	FilePath       string
	EncryptionKey  []byte
	DatabaseDSN    string
	MigrationsPath string
}

func New() *Config {
	key := []byte(getEnv("ENCRYPTION_KEY", ""))
	if len(key) == 0 {
		key = generateNewEncryptionKey()
	}

	return &Config{
		BaseURL:        "http://localhost:8080",
		ServerAddress:  ":8080",
		FilePath:       "",
		EncryptionKey:  key,
		DatabaseDSN:    "",
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://internal/app/storage/migrations/"),
	}
}

func generateNewEncryptionKey() []byte {
	randomGenerator := random.TrulyRandomGenerator{}
	randomKey, err := randomGenerator.GenerateRandomBytes(KeySize)
	if err != nil {
		randomKey = make([]byte, KeySize)
	}
	return randomKey
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
