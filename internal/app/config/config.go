package config

import (
	"os"
)

type Config struct {
	BaseURL       string
	ServerAddress string
}

func New() Config {
	return Config{
		ServerAddress: getServerAddress(),
		BaseURL:       getBaseURL(),
	}
}

func getServerAddress() string {
	v := os.Getenv("SERVER_ADDRESS")
	if v == "" {
		v = ":8080"
	}
	return v
}

func getBaseURL() string {
	v := os.Getenv("BASE_URL")
	if v == "" {
		v = "http://localhost:8080"
	}
	return v
}
