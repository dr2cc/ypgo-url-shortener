// Package config contains configuration for application.
package config

import (
	"crypto/aes"
	"flag"
	"os"

	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
)

const KeySize = 2 * aes.BlockSize //nolint:gomnd

type Config struct {
	BaseURL        string // base URL of the application
	ServerAddress  string // address the server will listen on
	FilePath       string // path to the file that will be used as storage
	DatabaseDSN    string // DSN for the database
	MigrationsPath string // path to the folder containing the migrations
	EncryptionKey  []byte // key used to encrypt and decrypt values
	EnableHTTPS    bool
}

// New creates new config with default values. It reads values from env and command line options.
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
		EnableHTTPS:    false,
	}
}

// generateNewEncryptionKey generates a random key of the specified KeySize.
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
	flag.BoolVar(&c.EnableHTTPS, "s", getEnv("ENABLE_HTTPS", "") == "true", "enable https")
}

// If the environment variable exists, return it, otherwise return the fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
