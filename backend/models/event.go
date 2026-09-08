package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	EventDate   time.Time `json:"event_date"`
	Location    string    `json:"location"`
	BannerImage string    `json:"banner_image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
