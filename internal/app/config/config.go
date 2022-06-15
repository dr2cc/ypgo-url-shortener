package config

import (
	"flag"
	"os"
)

type Config struct {
	BaseURL       string
	ServerAddress string
	FilePath      string
	flagSet       *flag.FlagSet
}

func New() *Config {
	return &Config{
		BaseURL:       "http://localhost:8080",
		ServerAddress: ":8080",
		FilePath:      "",
		flagSet:       nil,
	}
}

func (c *Config) Init() error {
	c.flagSet = flag.NewFlagSet("", flag.PanicOnError)
	c.flagSet.StringVar(&c.ServerAddress, "a", getEnv("SERVER_ADDRESS", ":8080"), "host to listen on")
	c.flagSet.StringVar(&c.BaseURL, "b", getEnv("BASE_URL", "http://localhost:8080"), "base url")
	c.flagSet.StringVar(&c.FilePath, "f", getEnv("FILE_STORAGE_PATH", ""), "file storage path")
	err := c.flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
