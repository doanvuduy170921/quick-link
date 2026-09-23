package http

type ShortenInput struct {
	URL         string `json:"url" binding:"required,url" example:"http://google.com"`
	CustomAlias string `json:"custom_alias" binding:"omitempty" example:"test"`
}
