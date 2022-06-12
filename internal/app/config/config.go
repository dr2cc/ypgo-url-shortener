package config

import (
	"fmt"
	"net"
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
	ln, err := net.Listen("tcp", v)

	if err != nil {
		fmt.Printf("Can't listen on port %q: %s", v, err)
		os.Exit(1)
	}

	_ = ln.Close()
	return v
}

func getBaseURL() string {
	v := os.Getenv("BASE_URL")
	if v == "" {
		v = "http://localhost:8080/"
	}
	return v
}
