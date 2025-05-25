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
	"github.com/gorilla/mux"
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
	db.AutoMigrate(&models.User{})

	rdb := utils.NewRedisClient()
	q := queue.NewQueue(rdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := consumer.Worker{Queue: q}
	go w.Start(ctx)

	handler := &handlers.RequestHandler{Queue: q}

	rr := mux.NewRouter()
	rr.Use(mockAuthMiddleware)
	rr.Use(requestTimeMiddleware)
	rr.HandleFunc("/requests", handler.CreateRequest).Methods("POST")
	rr.HandleFunc("/requests/{id}/status", handler.GetRequestStatus).Methods("GET")

	r := gin.Default()
	r.POST("/register", auth.Register(db))
	r.POST("/login", auth.Login(db, rdb))

	r.GET("/me", auth.AuthRequired(rdb), func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"user": username})
	})

	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("requestTime", time.Now().Format(time.RFC3339))
		c.Next()
	})

	r.Any("/requests", gin.WrapH(rr))
	r.Any("/requests/*any", gin.WrapH(rr))
	// r.POST("/requests", handler.CreateRequestGin)
	// r.GET("/requests/:id/status", handler.GetRequestStatusGin)

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

	// r.Run(":8080")

}

// Middleware для симуляции userID
func mockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "userID", "user-123")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware для добавления времени
func requestTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "requestTime", time.Now().Format(time.RFC3339))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
