package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	http.Handle("/", http.HandlerFunc(ShortenerHandler))
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var Urls = make(map[string]string)

func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getShortURL(w, r)
	case http.MethodPost:
		saveShortURL(w, r)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}

func saveShortURL(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	uh := hash(url)
	if uh == "" {
		http.Error(w, "cannot generate short url", 500)
		return
	}
	Urls[uh] = string(url)
	fmt.Println("added short url")
	fmt.Println(Urls)
}

func getShortURL(w http.ResponseWriter, r *http.Request) {
	uId := strings.TrimPrefix(r.URL.Path, "/")
	fmt.Println("parsed id ", uId)
	http.Redirect(w, r, Urls[uId], http.StatusTemporaryRedirect)
}

func hash(s []byte) string {
	h := fnv.New32a()
	_, err := h.Write(s)
	if err != nil {
		return ""
	}
	return strconv.Itoa(int(h.Sum32()))
}
