package models

import (
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"uniqueIndex" json:"user_id"`
	FullName         string    `gorm:"not null" json:"full_name"`
	Angkatan         string    `json:"angkatan"`
	Prodi            string    `json:"prodi"`
	PekerjaanSaatIni string    `json:"pekerjaan_saat_ini"`
	Perusahaan       string    `json:"perusahaan"`
	DomisiliKota     string    `json:"domisili_kota"`
	FotoProfil       string    `json:"foto_profil"`
	TotalLikes       int64     `gorm:"-" json:"total_likes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
