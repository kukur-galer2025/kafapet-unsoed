package controllers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"kafapet-backend/config"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/nfnt/resize"
)

// processAndSaveImage compresses & resizes an uploaded image, returns the public URL
func processAndSaveImage(c *fiber.Ctx, fieldKey string, fileIndex int, userID uint) (string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return "", err
	}

	files := form.File[fieldKey]
	if fileIndex >= len(files) {
		return "", fmt.Errorf("file index out of range")
	}

	file := files[fileIndex]

	// Ensure directory exists
	os.MkdirAll("./uploads/posts", os.ModePerm)

	// Open file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Try to decode as image for compression
	img, _, decodeErr := image.Decode(src)
	if decodeErr != nil {
		// Fallback: save raw file (e.g. GIF, WebP not decodable)
		ext := strings.ToLower(filepath.Ext(file.Filename))
		filename := fmt.Sprintf("%d-%d-%d%s", userID, time.Now().UnixNano(), fileIndex, ext)
		savePath := fmt.Sprintf("./uploads/posts/%s", filename)
		// Reset reader and save
		src.Close()
		src2, _ := file.Open()
		defer src2.Close()

		outFile, err := os.Create(savePath)
		if err != nil {
			return "", err
		}
		defer outFile.Close()

		buf := make([]byte, 1024*64)
		for {
			n, readErr := src2.Read(buf)
			if n > 0 {
				outFile.Write(buf[:n])
			}
			if readErr != nil {
				break
			}
		}
		return fmt.Sprintf("http://localhost:3000/uploads/posts/%s", filename), nil
	}

	// Resize (max width 1200px for good quality grid + lightbox)
	m := resize.Resize(1200, 0, img, resize.Lanczos3)

	// Save as compressed JPEG
	filename := fmt.Sprintf("%d-%d-%d.jpg", userID, time.Now().UnixNano(), fileIndex)
	savePath := fmt.Sprintf("./uploads/posts/%s", filename)

	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	err = jpeg.Encode(out, m, &jpeg.Options{Quality: 75})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://localhost:3000/uploads/posts/%s", filename), nil
}

// CreatePost creates a new post with up to 4 images
func CreatePost(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	content := c.FormValue("content")
	if content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content is required"})
	}

	var imageURLs []string

	// Handle multiple image uploads
	form, err := c.MultipartForm()
	if err == nil && form.File["images"] != nil {
		files := form.File["images"]
		if len(files) > 4 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Maximum 4 images allowed"})
		}

		for i := range files {
			url, err := processAndSaveImage(c, "images", i, userID)
			if err != nil {
				continue // Skip failed images
			}
			imageURLs = append(imageURLs, url)
		}
	}

	// Convert image URLs to JSON string
	imagesJSON := "[]"
	if len(imageURLs) > 0 {
		jsonBytes, _ := json.Marshal(imageURLs)
		imagesJSON = string(jsonBytes)
	}

	tagsRaw := c.FormValue("tags")
	var tagsArray []string
	if tagsRaw != "" {
		for _, tag := range strings.Split(tagsRaw, ",") {
			t := strings.TrimSpace(tag)
			if t != "" {
				if !strings.HasPrefix(t, "#") {
					t = "#" + t
				}
				tagsArray = append(tagsArray, t)
			}
		}
	}
	tagsJSON := "[]"
	if len(tagsArray) > 0 {
		jsonBytes, _ := json.Marshal(tagsArray)
		tagsJSON = string(jsonBytes)
	}

	post := models.Post{
		UserID:  userID,
		Content: content,
		Slug:    generateSlug(content),
		Images:  imagesJSON,
		Tags:    tagsJSON,
	}

	if err := config.DB.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create post"})
	}

	config.DB.Preload("User.Profile").First(&post, post.ID)

	return c.Status(fiber.StatusCreated).JSON(post)
}

func generateSlug(content string) string {
	// Take first 30 chars
	if len(content) > 30 {
		content = content[:30]
	}
	// Lowercase and replace non-alphanumeric with dash
	content = strings.ToLower(content)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(content, "-")
	slug = strings.Trim(slug, "-")

	// Append a random 6-character hex string
	rand.Seed(time.Now().UnixNano())
	randomSuffix := fmt.Sprintf("%06x", rand.Intn(16777215)) // max 6 hex chars

	if slug == "" {
		return "post-" + randomSuffix
	}
	return slug + "-" + randomSuffix
}

// GetPostBySlug fetches a single post by its slug or ID
func GetPostBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Slug is required"})
	}

	var post models.Post
	query := config.DB.Preload("User.Profile")
	
	// If slug is purely numeric, try matching ID as fallback for old posts
	if isNumeric(slug) {
		query = query.Where("slug = ? OR id = ?", slug, slug)
	} else {
		query = query.Where("slug = ?", slug)
	}

	if err := query.First(&post).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	return c.JSON(post)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// UpdatePost updates a post's content and images (only within 30 minutes of creation)
func UpdatePost(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	postID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	// Check ownership
	if post.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You can only edit your own posts"})
	}

	// Check 30-minute edit window
	if time.Since(post.CreatedAt) > 30*time.Minute {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Post can only be edited within 30 minutes of creation"})
	}

	// Update content
	newContent := c.FormValue("content")
	if newContent != "" {
		post.Content = newContent
	}

	// Handle existing images (images to keep)
	existingImagesStr := c.FormValue("existing_images") // JSON array of URLs to keep
	var keepImages []string
	if existingImagesStr != "" {
		json.Unmarshal([]byte(existingImagesStr), &keepImages)
	}

	// Parse old images to find which ones to delete from disk
	var oldImages []string
	json.Unmarshal([]byte(post.Images), &oldImages)

	for _, oldURL := range oldImages {
		found := false
		for _, keepURL := range keepImages {
			if oldURL == keepURL {
				found = true
				break
			}
		}
		if !found {
			// Delete from disk
			deleteImageFile(oldURL)
		}
	}

	// Handle new image uploads
	var newImageURLs []string
	form, formErr := c.MultipartForm()
	if formErr == nil && form.File["new_images"] != nil {
		files := form.File["new_images"]
		for i := range files {
			url, err := processAndSaveImage(c, "new_images", i, userID)
			if err != nil {
				continue
			}
			newImageURLs = append(newImageURLs, url)
		}
	}

	// Combine kept + new (max 4)
	allImages := append(keepImages, newImageURLs...)
	if len(allImages) > 4 {
		allImages = allImages[:4]
	}

	imagesJSON := "[]"
	if len(allImages) > 0 {
		jsonBytes, _ := json.Marshal(allImages)
		imagesJSON = string(jsonBytes)
	}
	post.Images = imagesJSON

	// Handle tags
	if tagsRaw := c.FormValue("tags"); tagsRaw != "" {
		var tagsArray []string
		for _, tag := range strings.Split(tagsRaw, ",") {
			t := strings.TrimSpace(tag)
			if t != "" {
				if !strings.HasPrefix(t, "#") {
					t = "#" + t
				}
				tagsArray = append(tagsArray, t)
			}
		}
		if len(tagsArray) > 0 {
			jsonBytes, _ := json.Marshal(tagsArray)
			post.Tags = string(jsonBytes)
		} else {
			post.Tags = "[]"
		}
	} else if c.Method() == fiber.MethodPut {
        // If it's a PUT request and tags are empty, we might want to clear them, 
        // but let's assume if it's sent empty it means no tags.
        // Wait, for FormData, if no tags are provided, it's empty string.
        post.Tags = "[]"
    }

	if err := config.DB.Save(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update post"})
	}

	config.DB.Preload("User.Profile").First(&post, post.ID)

	return c.JSON(fiber.Map{
		"message": "Post updated successfully",
		"data":    post,
	})
}

// DeletePost deletes a post and its associated data
func DeletePost(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	postID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	// Check ownership
	if post.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You can only delete your own posts"})
	}

	// Delete image files from disk
	var images []string
	json.Unmarshal([]byte(post.Images), &images)
	for _, imgURL := range images {
		deleteImageFile(imgURL)
	}

	// Delete associated likes and comments
	config.DB.Where("post_id = ?", postID).Delete(&models.Like{})
	config.DB.Where("post_id = ?", postID).Delete(&models.Comment{})

	// Delete the post (hard delete)
	config.DB.Unscoped().Delete(&post)

	return c.JSON(fiber.Map{
		"message": "Post deleted successfully",
	})
}

// deleteImageFile removes an uploaded image file from disk given its URL
func deleteImageFile(imageURL string) {
	if imageURL == "" {
		return
	}
	// Extract filename from URL like "http://localhost:3000/uploads/posts/filename.jpg"
	parts := strings.Split(imageURL, "/uploads/posts/")
	if len(parts) == 2 {
		filepath := "./uploads/posts/" + parts[1]
		os.Remove(filepath)
	}
}

// GetPosts returns all posts (optionally filtered by user_id)
func GetPosts(c *fiber.Ctx) error {
	var posts []models.Post

	query := config.DB.Preload("User.Profile").Preload("Likes").Preload("Comments.User.Profile").Order("created_at desc")

	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&posts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch posts"})
	}

	// Compute counts
	for i := range posts {
		posts[i].LikeCount = int64(len(posts[i].Likes))
		posts[i].CommentCount = int64(len(posts[i].Comments))
		// Clear the full likes array to reduce payload, keep only count
		posts[i].Likes = nil
	}

	return c.JSON(fiber.Map{
		"data": posts,
	})
}

// GetTrendingTags aggregates tags from posts to find the most popular ones
func GetTrendingTags(c *fiber.Ctx) error {
	var posts []models.Post
	// Fetch recent posts that might have tags (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	config.DB.Where("created_at > ?", thirtyDaysAgo).Find(&posts)

	tagCounts := make(map[string]int)

	for _, post := range posts {
		if post.Tags != "" && post.Tags != "[]" {
			var tags []string
			if err := json.Unmarshal([]byte(post.Tags), &tags); err == nil {
				for _, tag := range tags {
					tagCounts[tag]++
				}
			}
		}
	}

	// Sort tags by frequency
	type TagStat struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}

	var stats []TagStat
	for tag, count := range tagCounts {
		stats = append(stats, TagStat{Tag: tag, Count: count})
	}

	// Simple bubble sort for simplicity since list is usually small
	for i := 0; i < len(stats); i++ {
		for j := 0; j < len(stats)-i-1; j++ {
			if stats[j].Count < stats[j+1].Count {
				stats[j], stats[j+1] = stats[j+1], stats[j]
			}
		}
	}

	// Take top 5
	limit := 5
	if len(stats) < 5 {
		limit = len(stats)
	}

	return c.JSON(fiber.Map{
		"data": stats[:limit],
	})
}
