package controllers

import (
	"strconv"
	"kafapet-backend/config"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
)

// FollowUser creates a new connection
func FollowUser(c *fiber.Ctx) error {
	followerID := uint(c.Locals("user_id").(float64))

	followeeIDParam := c.Params("id")
	followeeID, err := strconv.ParseUint(followeeIDParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if followerID == uint(followeeID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You cannot follow yourself"})
	}

	// Check if followee exists
	var followee models.User
	if err := config.DB.First(&followee, followeeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Check if already following
	var existingConnection models.Connection
	result := config.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&existingConnection)
	if result.RowsAffected > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Already following this user"})
	}

	connection := models.Connection{
		FollowerID: followerID,
		FolloweeID: uint(followeeID),
	}

	if err := config.DB.Create(&connection).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to follow user"})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully followed user",
	})
}

// UnfollowUser deletes a connection
func UnfollowUser(c *fiber.Ctx) error {
	followerID := uint(c.Locals("user_id").(float64))

	followeeIDParam := c.Params("id")
	followeeID, err := strconv.ParseUint(followeeIDParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var connection models.Connection
	result := config.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&connection)
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not following this user"})
	}

	if err := config.DB.Delete(&connection).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unfollow user"})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully unfollowed user",
	})
}

// GetUserConnections retrieves follower and following counts/lists for a specific user
func GetUserConnections(c *fiber.Ctx) error {
	userIDParam := c.Params("id")
	var userID uint

	if userIDParam == "me" {
		userIDRaw := c.Locals("user_id")
		if userIDRaw == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		userID = uint(userIDRaw.(float64))
	} else {
		parsedID, err := strconv.ParseUint(userIDParam, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		userID = uint(parsedID)
	}

	var followers []models.Connection
	var following []models.Connection

	config.DB.Where("followee_id = ?", userID).Preload("Follower.Profile").Find(&followers)
	config.DB.Where("follower_id = ?", userID).Preload("Followee.Profile").Find(&following)

	return c.JSON(fiber.Map{
		"follower_count":  len(followers),
		"following_count": len(following),
		"followers":       followers,
		"following":       following,
	})
}
