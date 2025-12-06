package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHTTPServer_Shutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         "127.0.0.1:0", // случайный свободный порт
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// стартуем сервер
	go func() {
		_ = srv.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)

	// проверяем, что shutdown проходит без ошибок
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("server shutdown error: %v", err)
	}
}

