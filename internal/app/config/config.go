package config

import "os"

type Config struct {
	BaseUrl       string
	ServerAddress string
}

func New() Config {
	return Config{
		ServerAddress: getServerAddress(),
		BaseUrl:       getBaseUrl(),
	}
}

func getServerAddress() string {
	v := os.Getenv("SERVER_ADDRESS")
	if v == "" {
		v = "http://localhost:8080"
	}
	return v
}

func getBaseUrl() string {
	v := os.Getenv("BASE_URL")
	if v == "" {
		v = "http://localhost:8080/"
	}
	return v
}
