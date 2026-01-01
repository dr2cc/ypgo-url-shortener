// Package http_handlers contains functions that handles http requests
package handlers

import (
	"compress/flate"
	"compress/gzip"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Кука для iter14
const UserIDCookieName = "shortener-user-id"

// 01.01.2026 Как я теперь понимаю здесь должны быть обязательно только три❗ сущности:
// 🔸Handler struct  - главное назначение- передача запросов на уровень ниже --> service
// 🔸func NewHandler - конструктор сущности Handler
// 🔸func NewRouter (InitRoutes правильнее, это не конструктор) - описание всех обработчиков
// Остальное- в зависимости от функционала приложения.
// Не нужно все сносить сюда! Распределять по слоям!

type Handler struct {
	Mux     *chi.Mux             // router that we'll be using to handle our requests
	service *services.Shortener  // service that will contain main business logic
	crypto  crypto.Cryptographer // interface that we'll use to encrypt and decrypt values
}

// NewHandler creates a new instance of the Handler struct, initializes the chi mux, and sets the service and crypto fields
func NewHandler(service *services.Shortener, config *config.Config) *Handler {
	cryptographer := crypto.GCMAESCryptographer{Key: config.EncryptionKey, Random: service.Random}
	return &Handler{
		Mux:     chi.NewMux(),
		service: service,
		crypto:  &cryptographer,
	}
}

// NewRouter creates a new router, adds some middleware, and then adds some routes
func NewRouter(service *services.Shortener, ipChecker services.IPCheckerInterface, config *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service, config)

	r.Get("/{id}", h.Expand)
	r.Post("/", h.Shorten)
	r.Post("/api/shorten", h.ShortenAPI)
	//
	// здешний iter12
	// Добавьте новый хендлер POST /api/shorten/batch,
	// принимающий в теле запроса множество URL для сокращения в формате:
	r.Post("/api/shorten/batch", h.ShortenBatchAPI)
	//
	// 42 - iter14 (здешний iter9)
	// 	Добавьте в сервис функциональность аутентификации пользователя.

	// Сервис должен иметь хендлер GET /api/user/urls,
	// который сможет вернуть пользователю все когда-либо сокращённые им URL в формате:
	// [
	//     {
	//         "short_url": "http://...",
	//         "original_url": "http://..."
	//     },
	//     ...
	// ]
	r.Get("/api/user/urls", h.UserURLs)
	//
	// здешний (и новый в 42-й к.) iter10
	// Добавьте в сервис хендлер GET /ping,
	// который при запросе проверяет соединение с базой данных.
	// При успешной проверке хендлер должен вернуть HTTP-статус 200 OK, при неуспешной — 500 Internal Server Error.
	//
	r.Get("/ping", h.Ping)
	//
	// здешний iter14
	// Далее добавьте в сервис новый асинхронный хендлер DELETE /api/user/urls,
	// который принимает список идентификаторов сокращённых URL для удаления в формате:
	r.Delete("/api/user/urls", h.DeleteUrls)
	//
	r.Group(func(r chi.Router) {
		r.Use(FromTrustedSubnet(ipChecker))
		r.Get("/api/internal/stats", h.Stats)
	})

	return r
}

// getUserID gets the userID from the cookie.
func (h *Handler) getUserID(r *http.Request) string {
	encodedCookie, err := r.Cookie(UserIDCookieName)
	if err != nil {
		return h.service.GenerateNewUserID()
	}

	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
	if err != nil {
		return h.service.GenerateNewUserID()
	}

	decryptedUserID, err := h.crypto.Decrypt(decodedCookie)
	if err != nil {
		return h.service.GenerateNewUserID()
	}

	return string(decryptedUserID)
}

// Ping правильнее вынести в отдельный файл (01.01.2026)
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func FromTrustedSubnet(checkerInterface services.IPCheckerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fromTrustedSubnet, err := checkerInterface.IsRequestFromTrustedSubnet(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			if !fromTrustedSubnet {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// If the request body is gzipped, return a gzip reader, otherwise return the request body (default reader)
func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

// addEncryptedUserIDToCookie encrypts the userID and setting it as a cookie.
func (h *Handler) addEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
	encryptedUserID, err := h.crypto.Encrypt([]byte(userID))
	if err != nil {
		return err
	}

	encodedCookieValue := hex.EncodeToString(encryptedUserID)

	http.SetCookie(
		*w,
		&http.Cookie{
			Name:  UserIDCookieName,
			Value: encodedCookieValue,
		},
	)
	return nil
}
