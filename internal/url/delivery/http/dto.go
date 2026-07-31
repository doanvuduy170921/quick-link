package http

type ShortenInput struct {
	URL string `json:"url"`
	Exp int    `json:"exp"`
}
