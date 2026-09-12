package http

import (
	"errors"
	"github.com/doanvuduy170921/quick-link/internal/middleware"
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

	var user_id *int64
	uid, exist := c.Get(middleware.AuthorizationPayloadKey)
	if exist {
		id := uid.(int64)
		user_id = &id
	}

	if input.CustomAlias != "" && user_id == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errors.New("cannot use custom alias").Error()})
		return
	}
	code, err := h.useCase.GenerateKey(c.Request.Context(), input.URL, input.CustomAlias, user_id)
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
