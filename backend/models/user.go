package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash string         `json:"-"`
	Provider     string         `gorm:"type:varchar(50);default:'local'" json:"provider"`
	ProviderID   string         `gorm:"type:varchar(100)" json:"provider_id"`
	Role         string         `gorm:"default:'alumni'" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	
	Profile      Profile   `gorm:"foreignKey:UserID" json:"profile"`
}
