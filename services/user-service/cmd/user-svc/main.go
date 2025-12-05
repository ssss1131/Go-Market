package main

import (
	"GoUser/internal/http/handlers"
	jwtutil "GoUser/pkg/jwt"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfgpkg "GoUser/internal"
	"GoUser/internal/repo"
	"GoUser/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	migr "GoUser/internal/migrate"
)

func main() {
	cfg := cfgpkg.MustLoad()
	migr.Up(cfg.PGURL)

	db, err := gorm.Open(postgres.Open(cfg.PGURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	signer := jwtutil.NewSigner(cfg.JWTSecret, cfg.AccessTTL)

	usersRepo := repo.NewUsers(db)
	authSvc := service.NewAuthService(usersRepo, signer)
	authH := handlers.NewAuthHandler(authSvc)

	r := gin.Default()
	r.POST("/auth/register", authH.Register)
	r.POST("/auth/login", authH.Login)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("user-svc listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Println("user-svc stopped cleanly")
}
