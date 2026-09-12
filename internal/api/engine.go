package api

import (
	"context"
	clickevent "github.com/doanvuduy170921/quick-link/internal/click/event"
	clickrepo "github.com/doanvuduy170921/quick-link/internal/click/repository"
	"github.com/doanvuduy170921/quick-link/internal/middleware"
	http3 "github.com/doanvuduy170921/quick-link/internal/user/delivery/http"
	repository2 "github.com/doanvuduy170921/quick-link/internal/user/repository"
	usecase2 "github.com/doanvuduy170921/quick-link/internal/user/usecase"
	pkgJwt "github.com/doanvuduy170921/quick-link/pkg/jwt"
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

func NewServer(ctx context.Context, cfg *configs.Config, redisClient infrastructure.RedisClient, store db.Querier, jwtMgr *pkgJwt.JWTManager) *Server {
	s := &Server{
		app: gin.Default(),
		cfg: cfg,
	}
	s.SetUpRoutes(ctx, redisClient, store, jwtMgr)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.app.ServeHTTP(w, r)
}

func (s *Server) SetUpRoutes(ctx context.Context, redis infrastructure.RedisClient, store db.Querier, jwtMgr *pkgJwt.JWTManager) {

	clickRepo := clickrepo.NewClickRepository(store)

	clickTrack := clickevent.NewClickTracker(clickRepo, 1000)

	clickTrack.StartWorker(ctx, 5)
	urlRepo := repository.NewUrlRepository(store)
	urlUseCase := usecase.NewUseCase(urlRepo, redis)
	urlHandler := http2.NewURLHandler(urlUseCase, clickTrack)

	uRepo := repository2.NewUserRepository(store)
	uUseCase := usecase2.NewUserUseCase(uRepo, jwtMgr, 24*time.Hour)
	uHandler := http3.NewUserHandler(uUseCase)

	rateLimiter := middleware.NewRateLimiter(redis, 1*time.Minute, 10)
	s.app.POST("/shorten", rateLimiter.RateLimit(), middleware.OptionalAuthMiddleware(jwtMgr), urlHandler.ShortenURL)
	s.app.GET("/redirect/:code", urlHandler.Redirect)

	authGroup := s.app.Group("/api/v1/auth")
	{
		authGroup.POST("/register", uHandler.Register)
		authGroup.POST("/login", uHandler.Login)
	}

	// Protected Routes (Ví dụ API cần User Login - Dùng AuthMiddleware RSA)
	protectedGroup := s.app.Group("/api/v1").Use(middleware.AuthMiddleware(jwtMgr))
	{
		// Sau này thêm API GET /my-links vào đây
		_ = protectedGroup
	}
}
