package models

import "gorm.io/gorm"

type Request struct {
	gorm.Model
	ID        string `gorm:"primaryKey"`
	UserID    string
	CarModel  string
	Status    string
	CreatedAt string
}
