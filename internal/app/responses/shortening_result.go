package responses

type ShorteningResult struct {
	Result string `json:"result"`
}

type UsersShortUrl struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
