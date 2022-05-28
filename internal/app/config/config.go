package config

import "os"

type Config struct {
	Port string
	Host string
}

func New() Config {
	return Config{
		Port: Port(),
		Host: Host(),
	}
}

func Port() string {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8080"
	}

	return p
}

func Host() string {
	h := os.Getenv("HOST")
	if h == "" {
		h = "http://localhost"
	}

	return h
}
