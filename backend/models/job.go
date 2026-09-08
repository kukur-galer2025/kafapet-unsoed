package models

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"user"`
	Slug         string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"slug"`
	Title        string         `gorm:"not null" json:"title"`
	Company      string         `gorm:"not null" json:"company"`
	Location     string         `json:"location"`
	Description  string         `gorm:"type:text" json:"description"`
	Requirements string         `gorm:"type:text" json:"requirements"`
	JobType      string         `gorm:"default:'Full-time'" json:"job_type"`
	SalaryRange  string         `json:"salary_range"`
	ApplyLink    string         `json:"apply_link"`
	BannerImage  string         `json:"banner_image"`
	Status       string         `gorm:"default:'open'" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
