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
	"GoProduct/internal/http/middleware"
	migr "GoProduct/internal/migrate"
	"GoProduct/internal/repo"
	"GoProduct/internal/service"
	jwtutil "GoProduct/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := cfgpkg.MustLoad()
	migr.Up(cfg.PGURL)

	db, err := gorm.Open(postgres.Open(cfg.PGURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	verifier := jwtutil.NewVerifier(cfg.JWTSecret)

	productsRepo := repo.NewProducts(db)
	cartRepo := repo.NewCart(db)

	productSvc := service.NewProductService(productsRepo)
	cartSvc := service.NewCartService(cartRepo, productSvc)

	productH := handlers.NewProductHandler(productSvc)
	cartH := handlers.NewCartHandler(cartSvc)

	r := gin.Default()

	products := r.Group("/products", middleware.AuthRequired(verifier))
	{
		products.GET("/", productH.List)
		products.GET("/:id", productH.Get)

		write := products.Group("", middleware.RequireActive())
		{
			write.POST("/", productH.Create)
			write.PUT("/:id", productH.Update)
			write.DELETE("/:id", productH.Delete)
		}

		cart := products.Group("/cart", middleware.RequireActive())
		{
			cart.GET("/", cartH.GetCart)
			cart.POST("/", cartH.AddItem)
			cart.PUT("/:product_id", cartH.UpdateItem)
			cart.DELETE("/:product_id", cartH.DeleteItem)
		}
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
