package controllers

import (
	"kafapet-backend/config"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
)

// ToggleLike toggles a like on a post (like if not liked, unlike if already liked)
func ToggleLike(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	postID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	// Check if post exists
	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	// Check if user already liked this post
	var existingLike models.Like
	result := config.DB.Where("user_id = ? AND post_id = ?", userID, uint(postID)).First(&existingLike)

	if result.RowsAffected > 0 {
		// Unlike: remove the like
		config.DB.Delete(&existingLike)

		// Get updated count
		var count int64
		config.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&count)

		return c.JSON(fiber.Map{
			"liked":      false,
			"like_count": count,
			"message":    "Like removed",
		})
	}

	// Like: create new like
	like := models.Like{
		UserID: userID,
		PostID: uint(postID),
	}

	if err := config.DB.Create(&like).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to like post"})
	}

	// Get updated count
	var count int64
	config.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&count)

	return c.JSON(fiber.Map{
		"liked":      true,
		"like_count": count,
		"message":    "Post liked",
	})
}

// GetLikes returns the list of users who liked a post
func GetLikes(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var likes []models.Like
	config.DB.Preload("User.Profile").Where("post_id = ?", postID).Find(&likes)

	return c.JSON(fiber.Map{
		"data":  likes,
		"count": len(likes),
	})
}
