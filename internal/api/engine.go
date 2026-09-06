package api

import (
	"context"
	clickevent "github.com/doanvuduy170921/quick-link/internal/click/event"
	clickrepo "github.com/doanvuduy170921/quick-link/internal/click/repository"
	"github.com/doanvuduy170921/quick-link/internal/middleware"
	"net/http"
	"time"

	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/infrastructure"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	http2 "github.com/doanvuduy170921/quick-link/internal/url/delivery/http"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"github.com/gin-gonic/gin"
)

type Server struct {
	app *gin.Engine
	cfg *configs.Config
}

func NewServer(ctx context.Context, cfg *configs.Config, redisClient infrastructure.RedisClient, store db.Querier) *Server {
	s := &Server{
		app: gin.Default(),
		cfg: cfg,
	}
	s.SetUpRoutes(ctx, redisClient, store)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.app.ServeHTTP(w, r)
}

func (s *Server) SetUpRoutes(ctx context.Context, redis infrastructure.RedisClient, store db.Querier) {

	clickRepo := clickrepo.NewClickRepository(store)

	clickTrack := clickevent.NewClickTracker(clickRepo, 1000)

	clickTrack.StartWorker(ctx, 5)
	urlRepo := repository.NewUrlRepository(store)
	urlUseCase := usecase.NewUseCase(urlRepo, redis)
	urlHandler := http2.NewURLHandler(urlUseCase, clickTrack)

	rateLimiter := middleware.NewRateLimiter(redis, 1*time.Minute, 10)
	s.app.POST("/shorten", rateLimiter.RateLimit(), urlHandler.ShortenURL)
	s.app.GET("/redirect/:code", urlHandler.Redirect)
}
