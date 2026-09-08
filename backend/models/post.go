package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Slug      string         `gorm:"type:varchar(255);uniqueIndex" json:"slug"`
	Images    string         `gorm:"type:text" json:"images"` // JSON array of image URLs e.g. ["url1","url2"]
	Tags      string         `gorm:"type:text" json:"tags"`   // JSON array of strings e.g. ["loker", "event"]
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User         User      `gorm:"foreignKey:UserID" json:"user"`
	Likes        []Like    `gorm:"foreignKey:PostID" json:"likes,omitempty"`
	LikeCount    int64     `gorm:"-" json:"like_count"`
	Comments     []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
	CommentCount int64     `gorm:"-" json:"comment_count"`
}
