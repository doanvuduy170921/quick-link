package http

type ShortenInput struct {
	URL string `json:"url" binding:"required,url"`
}
