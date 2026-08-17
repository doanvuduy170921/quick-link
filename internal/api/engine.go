package api

import (
	"errors"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	http2 "github.com/doanvuduy170921/quick-link/internal/url/delivery/http"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Server struct {
	app *gin.Engine
	cfg *configs.Config
}

func NewServer(cfg *configs.Config, redisClient infrastructure.RedisClient, store db.Querier) *Server {
	s := &Server{
		app: gin.Default(),
		cfg: cfg,
	}
	s.SetUpRoutes(redisClient, store)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv := &http.Server{
		Addr:    ":" + s.cfg.Server.Port,
		Handler: s.app,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}

func (s *Server) SetUpRoutes(redis infrastructure.RedisClient, store db.Querier) {
	urlRepo := repository.NewUrlRepository(store)
	urlUseCase := usecase.NewUseCase(urlRepo, redis)
	urlHandler := http2.NewURLHandler(urlUseCase)

	s.app.POST("/shorten", urlHandler.ShortenURL)
	s.app.GET("/redirect/:code", urlHandler.Redirect)
}
