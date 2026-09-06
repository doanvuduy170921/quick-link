package boot

import (
	"context"
	"errors"
	"fmt"
	"github.com/doanvuduy170921/quick-link/configs"
	"github.com/doanvuduy170921/quick-link/internal/api"
	"github.com/doanvuduy170921/quick-link/internal/container"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func RunServer(cfg *configs.Config, cnt *container.Container) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	server := api.NewServer(ctx, cfg, cnt.Redis, cnt.Query)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      server,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Println("shutting down...")
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	return srv.Shutdown(shutdownCtx)

}
