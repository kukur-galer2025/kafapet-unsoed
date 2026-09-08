package controllers

import (
	"kafapet-backend/config"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
)

func GetEvents(c *fiber.Ctx) error {
	var events []models.Event
	if err := config.DB.Order("event_date desc").Find(&events).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
	}
	return c.JSON(fiber.Map{"data": events})
}

func CreateEvent(c *fiber.Ctx) error {
	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := config.DB.Create(&event).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create event"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Event created successfully",
		"data":    event,
	})
}
