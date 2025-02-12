package models

import "gorm.io/gorm"

var DB *gorm.DB

type Book struct {
	ID     int    `json:"id" gorm:"primaryKey"`
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
	Year   int    `json:"year" binding:"required"`
}
