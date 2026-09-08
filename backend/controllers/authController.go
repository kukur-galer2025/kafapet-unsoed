package controllers

import (
	"kafapet-backend/config"
	"kafapet-backend/models"
	"kafapet-backend/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Cek apakah email sudah terdaftar
	var existingUser models.User
	if err := config.DB.Where("email = ?", strings.ToLower(input.Email)).First(&existingUser).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Email already registered"})
	}

	// Hash Password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create User
	user := models.User{
		Email:        strings.ToLower(input.Email),
		PasswordHash: hashedPassword,
		Role:         "alumni",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// Create Profile
	profile := models.Profile{
		UserID:   user.ID,
		FullName: input.FullName,
	}
	config.DB.Create(&profile)

	return c.Status(201).JSON(fiber.Map{
		"message": "User registered successfully",
		"user_id": user.ID,
	})
}

func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Cari user
	var user models.User
	if err := config.DB.Where("email = ?", strings.ToLower(input.Email)).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Cek password
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
	})
}
