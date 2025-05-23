package main

import (
	"net/http"
	"premium_cars_app/internal/auth"
	"premium_cars_app/pkg/models"
	"premium_cars_app/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
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
