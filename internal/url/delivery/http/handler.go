package http

import (
	url_err "github.com/doanvuduy170921/quick-link/internal/url/error"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
)

type URLHandler struct {
	useCase      usecase.UseCase
	clickTracker ClickTracker
}

type ClickTracker interface {
	Track(code string)
}

func NewURLHandler(useCase usecase.UseCase, clickTracker ClickTracker) *URLHandler {
	return &URLHandler{
		useCase:      useCase,
		clickTracker: clickTracker,
	}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var input ShortenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code, err := h.useCase.GenerateKey(c.Request.Context(), input.URL)
	if err != nil {
		url_err.HandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": "Shorten URL successfully!",
	})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.useCase.Redirect(c.Request.Context(), code)
	if err != nil {
		url_err.HandlerError(c, err)
		return
	}
	h.clickTracker.Track(code)

	c.Redirect(http.StatusFound, url)
}
