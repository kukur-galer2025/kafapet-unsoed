package main

import (
	"log"
	"os"

	"kafapet-backend/config"
	"kafapet-backend/controllers"
	"kafapet-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file (might be using system env vars)")
	}

	// Connect to Database
	config.ConnectDB()

	// Initialize OAuth
	controllers.InitOauthConfig()

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:4321", // Sesuaikan dengan port frontend Astro
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))

	// Setup API Routes
	routes.SetupRoutes(app)

	// Serve Static Files (For Uploaded Images)
	app.Static("/uploads", "./uploads")

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to KAFAPET API",
			"status":  "success",
		})
	})

	// Setup Server Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server is running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
