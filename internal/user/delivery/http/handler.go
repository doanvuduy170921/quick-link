package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/doanvuduy170921/quick-link/internal/user/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	useCase usecase.UserUseCase
}

func NewUserHandler(useCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		useCase: useCase,
	}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce      json
// @Param        request body RegisterRequest true "Registration info"
// @Success      201 {object} UserResponse
// @Failure      400 {object} map[string]string
// @Failure      409 {object} map[string]string "email already registered"
// @Router       /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.useCase.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

// Login godoc
// @Summary      Login and get JWT access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Credentials"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string "invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.useCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: token,
	})
}
