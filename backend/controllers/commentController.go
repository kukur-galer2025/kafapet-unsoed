package controllers

import (
	"kafapet-backend/config"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
)

// CreateComment creates a new comment on a post (requires auth)
func CreateComment(c *fiber.Ctx) error {
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

	// Parse body
	type CommentInput struct {
		Content string `json:"content"`
	}

	var input CommentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	comment := models.Comment{
		UserID:  userID,
		PostID:  uint(postID),
		Content: input.Content,
	}

	if err := config.DB.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create comment"})
	}

	// Preload user data
	config.DB.Preload("User.Profile").First(&comment, comment.ID)

	// Get updated count
	var count int64
	config.DB.Model(&models.Comment{}).Where("post_id = ?", postID).Count(&count)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":          comment,
		"comment_count": count,
	})
}

// GetComments returns all comments for a post (public)
func GetComments(c *fiber.Ctx) error {
	postID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var comments []models.Comment
	config.DB.Preload("User.Profile").Where("post_id = ?", postID).Order("created_at asc").Find(&comments)

	return c.JSON(fiber.Map{
		"data":  comments,
		"count": len(comments),
	})
}

// DeleteComment deletes a comment (only by the comment owner)
func DeleteComment(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	commentID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid comment ID"})
	}

	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Comment not found"})
	}

	// Only the owner can delete their comment
	if comment.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You can only delete your own comments"})
	}

	postID := comment.PostID
	config.DB.Delete(&comment)

	// Get updated count
	var count int64
	config.DB.Model(&models.Comment{}).Where("post_id = ?", postID).Count(&count)

	return c.JSON(fiber.Map{
		"message":       "Comment deleted",
		"comment_count": count,
	})
}
