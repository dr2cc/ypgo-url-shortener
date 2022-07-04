package responses

type ShorteningResult struct {
	Result string `json:"result"`
}

type UsersShortURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
type ShorteningBatchResult struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
