// Package responses contains structs of responses send in http endpoints.
package responses

// ShorteningResult is response with shortened url.
type ShorteningResult struct {
	Result string `json:"result"`
}

// UsersShortURL is url that was shortened by user.
type UsersShortURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ShorteningBatchResult is shortening result of batch operation.
type ShorteningBatchResult struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
