// Package config contains configuration for application.
package config

import (
	"crypto/aes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/belamov/ypgo-url-shortener/internal/app/services/random"
)

const KeySize = 2 * aes.BlockSize //nolint:gomnd

type Config struct {
	BaseURL        string `json:"base_url"`
	ServerAddress  string `json:"server_address"`
	FilePath       string `json:"file_storage_path"`
	DatabaseDSN    string `json:"database_dsn"`
	MigrationsPath string
	ConfigPath     string
	EncryptionKey  []byte
	EnableHTTPS    bool `json:"enable_https"`
}

// New creates new config with default values. It reads values from env and command line options.
func New() *Config {
	key := []byte(getEnv("ENCRYPTION_KEY", ""))
	if len(key) == 0 {
		key = generateNewEncryptionKey()
	}

	return &Config{
		BaseURL:        "",
		ServerAddress:  "",
		FilePath:       "",
		EncryptionKey:  key,
		DatabaseDSN:    "",
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://internal/app/storage/migrations/"),
		EnableHTTPS:    false,
		ConfigPath:     "",
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

func (c *Config) Init() error {
	flag.StringVar(&c.ServerAddress, "a", getEnv("SERVER_ADDRESS", ""), "host to listen on")
	flag.StringVar(&c.BaseURL, "b", getEnv("BASE_URL", ""), "base url")
	flag.StringVar(&c.FilePath, "f", getEnv("FILE_STORAGE_PATH", ""), "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", getEnv("DATABASE_DSN", ""), "database dsn for connecting to postgres")
	flag.StringVar(&c.ConfigPath, "c", getEnv("CONFIG", ""), "config path")
	flag.BoolVar(&c.EnableHTTPS, "s", getEnv("ENABLE_HTTPS", "") == "true", "enable https")

	flag.Parse()

	if c.ConfigPath != "" {
		return c.loadEmptyValuesFromFile(c.ConfigPath)
	}

	if c.ServerAddress == "" {
		c.ServerAddress = ":8080"
	}

	if c.BaseURL == "" {
		c.BaseURL = "http://localhost:8080"
	}

	return nil
}

//nolint:cyclop
func (c *Config) loadEmptyValuesFromFile(configPath string) error {
	configFromFile, err := c.parseConfigFile(configPath)
	if err != nil {
		return err
	}

	if c.BaseURL == "" && configFromFile.BaseURL != "" {
		c.BaseURL = configFromFile.BaseURL
	}

	if c.ServerAddress == "" && configFromFile.ServerAddress != "" {
		c.ServerAddress = configFromFile.ServerAddress
	}

	if c.FilePath == "" && configFromFile.FilePath != "" {
		c.FilePath = configFromFile.FilePath
	}

	if c.DatabaseDSN == "" && configFromFile.DatabaseDSN != "" {
		c.DatabaseDSN = configFromFile.DatabaseDSN
	}

	if !c.EnableHTTPS && configFromFile.EnableHTTPS {
		c.EnableHTTPS = configFromFile.EnableHTTPS
	}

	return nil
}

func (c *Config) parseConfigFile(configPath string) (Config, error) {
	f, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config file not found at: %s", configPath)
		}
		return Config{}, err
	}

	configFromFile := Config{}

	err = json.Unmarshal(f, &configFromFile)
	return configFromFile, err
}

// If the environment variable exists, return it, otherwise return the fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
