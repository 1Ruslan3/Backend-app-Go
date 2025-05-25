package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"premium_cars_app/internal/auth"
	"premium_cars_app/internal/consumer"
	"premium_cars_app/internal/handlers"
	"premium_cars_app/internal/queue"
	"premium_cars_app/pkg/models"
	"premium_cars_app/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		"localhost",
		"postgres",
		"postgres",
		"premiumcars",
		"5433",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to PostgreSQL")
	}
	db.AutoMigrate(&models.User{}, &models.Request{})

	rdb := utils.NewRedisClient()
	q := queue.NewQueue(rdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := consumer.Worker{
		Queue: q,
		DB:    db,
	}
	go w.Start(ctx)

	handler := &handlers.RequestHandler{Queue: q}

	r := gin.Default()

	r.POST("/register", auth.Register(db))
	r.POST("/login", auth.Login(db, rdb))

	protected := r.Group("/")
	protected.Use(auth.AuthRequired(rdb))
	{
		protected.GET("/me", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{"user": username})
		})

		protected.POST("/requests", handler.CreateRequestGin)
		protected.GET("/requests/:id/status", handler.GetRequestStatusGin)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down...")

		cancel()

		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelTimeout()

		if err := srv.Shutdown(ctxTimeout); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Println("Server is running on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	// r.GET("/me", auth.AuthRequired(rdb), func(c *gin.Context) {
	// 	username, _ := c.Get("username")
	// 	c.JSON(http.StatusOK, gin.H{"user": username})
	// })

	// rr := mux.NewRouter()
	// rr.Use(mockAuthMiddleware)
	// rr.Use(requestTimeMiddleware)
	// rr.HandleFunc("/requests", handler.CreateRequest).Methods("POST")
	// rr.HandleFunc("/requests/{id}/status", handler.GetRequestStatus).Methods("GET")

	// r.Use(func(c *gin.Context) {
	// 	c.Set("userID", "user-123")
	// 	c.Set("requestTime", time.Now().Format(time.RFC3339))
	// 	c.Next()
	// })

	// r.Any("/requests", gin.WrapH(rr))
	// r.Any("/requests/*any", gin.WrapH(rr))
	// r.POST("/requests", handler.CreateRequestGin)
	// r.GET("/requests/:id/status", handler.GetRequestStatusGin)

	// r.Run(":8080")

}
