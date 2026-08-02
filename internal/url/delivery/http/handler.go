package http

import (
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
)

type URLHandler struct {
	useCase usecase.UseCase
}

func NewURLHandler(useCase usecase.UseCase) *URLHandler {
	return &URLHandler{
		useCase: useCase,
	}

}

func (h *URLHandler) ShortenURl(c *gin.Context) {
	var input ShortenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code, err := h.useCase.GenerateKey(c.Request.Context(), input.URL, input.Exp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"Message": "Shorten URL successfully!",
	})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.useCase.Redirect(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusMovedPermanently, url)

}
