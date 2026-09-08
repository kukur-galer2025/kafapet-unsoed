package controllers

import (
	"kafapet-backend/config"
	"kafapet-backend/models"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetAlumniList(c *fiber.Ctx) error {
	var profiles []models.Profile
	
	// Pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Search parameters
	search := c.Query("search", "")

	query := config.DB.Model(&models.Profile{})
	
	if search != "" {
		query = query.Where("full_name LIKE ? OR angkatan LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch alumni"})
	}

	return c.JSON(fiber.Map{
		"data":  profiles,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64) // JWT claims parsing returns float64 for numbers

	var input models.Profile
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var profile models.Profile
	if err := config.DB.Where("user_id = ?", uint(userID)).First(&profile).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Profile not found"})
	}

	// Update allowed fields
	if input.FullName != "" {
		profile.FullName = input.FullName
	}
	profile.Angkatan = input.Angkatan
	profile.Prodi = input.Prodi
	profile.PekerjaanSaatIni = input.PekerjaanSaatIni
	profile.Perusahaan = input.Perusahaan
	profile.DomisiliKota = input.DomisiliKota

	if err := config.DB.Save(&profile).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update profile"})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"data":    profile,
	})
}

// GetSuggestedAlumni returns top alumni based on total likes on their posts
func GetSuggestedAlumni(c *fiber.Ctx) error {
	var profiles []models.Profile
	
	// Try to get user_id if logged in, to exclude them
	userID := c.Locals("user_id")
	
	query := config.DB.Model(&models.Profile{}).
		Select("profiles.*, COUNT(likes.id) as total_likes").
		Joins("LEFT JOIN posts ON posts.user_id = profiles.user_id").
		Joins("LEFT JOIN likes ON likes.post_id = posts.id").
		Group("profiles.id").
		Order("COUNT(likes.id) DESC").
		Limit(3)

	if userID != nil {
		query = query.Where("profiles.user_id != ?", userID)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch suggested alumni"})
	}

	return c.JSON(fiber.Map{
		"data": profiles,
	})
}

// UpdateAvatar handles profile picture uploads
func UpdateAvatar(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	// Get the file from form-data
	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to upload avatar"})
	}

	// Make sure directory exists (uploads/avatars)
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create directory"})
	}

	filePath := uploadDir + "/" + file.Filename
	
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save avatar"})
	}

	var profile models.Profile
	if err := config.DB.Where("user_id = ?", uint(userID)).First(&profile).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Profile not found"})
	}

	// Update FotoProfil URL
	profile.FotoProfil = "http://localhost:3000/uploads/avatars/" + file.Filename

	if err := config.DB.Save(&profile).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update profile picture"})
	}

	return c.JSON(fiber.Map{
		"message": "Avatar updated successfully",
		"data":    profile,
	})
}

// GetAlumniProfile returns a single alumni's public profile with stats
func GetAlumniProfile(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var profile models.Profile
	if err := config.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Alumni not found"})
	}

	// Get total likes
	var totalLikes int64
	config.DB.Model(&models.Like{}).
		Joins("JOIN posts ON posts.id = likes.post_id").
		Where("posts.user_id = ?", userID).
		Count(&totalLikes)

	// Get follower/following counts
	var followerCount int64
	var followingCount int64
	config.DB.Model(&models.Connection{}).Where("followee_id = ?", userID).Count(&followerCount)
	config.DB.Model(&models.Connection{}).Where("follower_id = ?", userID).Count(&followingCount)

	// Get recent posts
	var posts []models.Post
	config.DB.Where("user_id = ?", userID).
		Preload("User.Profile").
		Preload("Likes").
		Preload("Comments").
		Order("created_at desc").
		Limit(10).
		Find(&posts)

	for i := range posts {
		posts[i].LikeCount = int64(len(posts[i].Likes))
		posts[i].CommentCount = int64(len(posts[i].Comments))
		posts[i].Likes = nil
	}

	return c.JSON(fiber.Map{
		"profile":         profile,
		"total_likes":     totalLikes,
		"follower_count":  followerCount,
		"following_count": followingCount,
		"posts":           posts,
	})
}
