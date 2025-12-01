package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfgpkg "GoProduct/internal"
	"GoProduct/internal/http/handlers"
	"GoProduct/internal/repo"
	"GoProduct/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	migr "GoProduct/internal/migrate"
)

func main() {
	cfg := cfgpkg.MustLoad()
	migr.Up(cfg.PGURL)

	db, err := gorm.Open(postgres.Open(cfg.PGURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	productsRepo := repo.NewProducts(db)
	productSvc := service.NewProductService(productsRepo)
	productH := handlers.NewProductHandler(productSvc)

	r := gin.Default()

	products := r.Group("/products")
	{
		products.POST("/", productH.Create)
		products.GET("/", productH.List)
		products.GET("/:id", productH.Get)
		products.PUT("/:id", productH.Update)
		products.DELETE("/:id", productH.Delete)
	}

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("product-svc listening on %s", cfg.HTTPAddr)
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

	log.Println("product-svc stopped cleanly")
}
