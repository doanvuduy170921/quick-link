package boot

import (
	"context"
	"errors"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/api"
	"github.com/doanvuduy170921/quick-link/internal/container"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunServer(cfg *configs.Config, cnt *container.Container) {
	server := api.NewServer(cfg, cnt.Redis, cnt.Query)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      server,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exiting")
}
