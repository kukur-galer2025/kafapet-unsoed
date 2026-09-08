package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user"`
	Slug        string         `gorm:"type:varchar(255);uniqueIndex" json:"slug"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Content     string         `gorm:"type:longtext;not null" json:"content"`
	CoverImage  string         `gorm:"type:varchar(512)" json:"cover_image"`
	Tags        string         `gorm:"type:text" json:"tags"` // JSON array
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
