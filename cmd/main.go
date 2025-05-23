package main

import (
	"fmt"
	"net/http"
	"premium_cars_app/internal/auth"
	"premium_cars_app/pkg/models"
	"premium_cars_app/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

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

	r := gin.Default()
	r.POST("/register", auth.Register(db))
	r.POST("/login", auth.Login(db, rdb))

	r.GET("/me", auth.AuthRequired(rdb), func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"user": username})
	})

	r.Run(":8080")
}
