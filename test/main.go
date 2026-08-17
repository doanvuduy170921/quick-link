package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	fmt.Println(w, "Hello World")
}
func main() {
	r := gin.Default()

	r.GET("/hello-word", func(c *gin.Context) {
		log.Println("Start with task slow request , graceful shutdowns")
		time.Sleep(10 * time.Second)

		select {
		case <-c.Request.Context().Done():
			log.Println("Request cancelled due to graceful shutdowns")
			return
		default:
			log.Println("End with task slow request , graceful shutdowns")
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello World",
			})
			return
		}

	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Printf("Listening and serving HTTP on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	c := <-quit
	log.Println("Got signal:", c)
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exiting gracefully")

}
