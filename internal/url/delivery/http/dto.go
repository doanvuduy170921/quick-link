package http

type ShortenInput struct {
	URL         string `json:"url" binding:"required,url"`
	CustomAlias string `json:"custom_alias" binding:"omitempty"`
}
