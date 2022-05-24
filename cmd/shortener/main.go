package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(ShortenerHandler))

	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	// проверяем, каким методом получили запрос
	switch r.Method {
	// если методом POST
	case http.MethodGet:
		getShortUrl(w, r)
	case http.MethodPost:
		saveShortUrl(w, r)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}

func saveShortUrl(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func getShortUrl(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTemporaryRedirect)

}
