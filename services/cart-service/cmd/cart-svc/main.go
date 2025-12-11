package main

import (
	"GoCart/internal"
	"GoCart/internal/http/handlers"
	"GoCart/internal/http/middleware"
	"GoCart/internal/migrate"
	"GoCart/internal/repo"
	"GoCart/internal/service"
	jwt "GoProduct/pkg/jwt"

	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	cfg := internal.MustLoad()
	migrate.Up(cfg.PGURL)

	db, err := gorm.Open(postgres.Open(cfg.PGURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	cartRepo := repo.NewCartRepository(db)
	productClient := service.NewProductHTTPClient(cfg.ProductURL)
	cartSvc := service.NewCartService(cartRepo, productClient)
	cartH := handlers.NewCartHandler(cartSvc)

	verifier := jwt.NewVerifier(cfg.JWTSecret)

	r := gin.Default()

	r.Use(middleware.AuthRequired(verifier))

	cart := r.Group("/cart")
	{
		cart.GET("/", cartH.GetCart)
		cart.POST("/", middleware.RequireActive(), cartH.AddItem)
		cart.PUT("/:product_id", middleware.RequireActive(), cartH.UpdateItem)
		cart.DELETE("/:product_id", middleware.RequireActive(), cartH.DeleteItem)
	}

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("cart-service listening on %s", cfg.HTTPAddr)
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

	sqlDB, _ := db.DB()
	sqlDB.Close()

	log.Println("cart-service stopped cleanly")
}
