package config

import (
	"fmt"
	"log"
	"os"

	"kafapet-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbName)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	DB = db
	log.Println("Database connection successfully opened")

	// Fix legacy empty slugs before migrating (to prevent unique constraint error)
	DB.Exec("UPDATE jobs SET slug = CONCAT('job-legacy-', id) WHERE slug = '' OR slug IS NULL")

	// Auto Migrate Models
	err = DB.AutoMigrate(
		&models.User{}, 
		&models.Profile{},
		&models.Job{},
		&models.Event{},
		&models.Post{},
		&models.Like{},
		&models.Comment{},
		&models.Connection{},
		&models.Article{},
	)
	if err != nil {
		log.Fatal("Failed to auto migrate: ", err)
	}
	log.Println("Database Migrated")
}
