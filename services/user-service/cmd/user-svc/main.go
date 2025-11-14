package main

import (
	"GoMarket/internal/http/handlers"
	jwtutil "GoMarket/pkg/jwt"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfgpkg "GoMarket/internal"
	"GoMarket/internal/repo"
	"GoMarket/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := cfgpkg.MustLoad()

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
	}() // запускаем сервер конкурентно чтобы не блокировать исполнение нижнего кода(типа многопоточность)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // ждем такой сигнал
	<-quit                                               // блокаем нижний код пока не придёт уведомление(код выше)

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
