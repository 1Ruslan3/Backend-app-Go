package models

import (
	"time"

	"gorm.io/gorm"
)

type Request struct {
	gorm.Model
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	CarModel  string    `json:"car_model"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
