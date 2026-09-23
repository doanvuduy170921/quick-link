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
	ClickTracker ClickTracker
}

type ClickTracker interface {
	Track(code string)
}

func NewURLHandler(useCase usecase.UseCase, ClickTracker ClickTracker) *URLHandler {
	return &URLHandler{
		useCase:      useCase,
		ClickTracker: ClickTracker,
	}
}

// ShortenURL godoc
// @Summary Create a short URL
// @Description Shortens a long URL. Anonymous users get a random code (user_id = null).
// @Description Authenticated users may pass a custom_alias; anonymous users passing custom_alias get 401.
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        request body ShortenInput true "URL to shorten"
// @Success      200 {object} map[string]string "code, message"
// @Failure      400 {object} map[string]string "validation error / malformed body"
// @Failure      401 {object} map[string]string "custom_alias requires login"
// @Failure      500 {object} map[string]string "internal error"
// @Security     BearerAuth
// @Router       /shorten [post]
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

// Redirect godoc
// @Summary Redirect to original URL
// @Description  Redirects to the original URL for a given short code. Tracks click asynchronously.
// @Tags         url
// @Produce      json
// @Param        code path string true "Short code"
// @Success      302 "Redirects to original URL"
// @Failure      404 {object} map[string]string "code not found or expired"
// @Failure      500 {object} map[string]string "internal error"
// @Router       /redirect/{code} [get]
func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.useCase.Redirect(c.Request.Context(), code)
	if err != nil {
		url_err.HandlerError(c, err)
		return
	}
	h.ClickTracker.Track(code)

	c.Redirect(http.StatusFound, url)
}
