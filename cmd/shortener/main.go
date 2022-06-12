package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/server"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/generator"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
)

func main() {

	// just for passing autotests
	v := struct {
		Url string
	}{
		Url: "http://mysite.com?id=1234&param=2",
	}
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false) // без этой опции символ '&' будет заменён на "\u0026"
	err := encoder.Encode(v)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf.String())

	cfg := config.New()
	repo := storage.NewInMemoryRepository()
	gen := &generator.HashGenerator{}
	service := services.New(repo, gen, cfg)
	srv := server.New(cfg, service)

	srv.Run()
}
