package api

import (
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	http2 "github.com/doanvuduy170921/quick-link/internal/url/delivery/http"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	app *gin.Engine
	cfg *configs.Config
}

func NewServer(cfg *configs.Config, redisClient infrastructure.RedisClient) *Server {
	s := &Server{
		app: gin.Default(),
		cfg: cfg,
	}
	s.SetUpRoutes(redisClient)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.app.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	return s.app.Run(":" + s.cfg.Port)
}

func (s *Server) SetUpRoutes(redis infrastructure.RedisClient) {
	urlRepo := repository.NewUrlRepository(redis)
	urlUseCase := usecase.NewUseCase(urlRepo)
	urlHandler := http2.NewURLHandler(urlUseCase)

	s.app.POST("/shorten", urlHandler.ShortenURl)
	s.app.GET("/redirect/:code", urlHandler.Redirect)
}
