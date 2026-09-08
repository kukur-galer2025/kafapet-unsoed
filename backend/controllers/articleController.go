package controllers

import (
	"fmt"
	"kafapet-backend/config"
	"kafapet-backend/models"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// generateArticleSlug creates a URL-safe slug from title
func generateArticleSlug(title string) string {
	s := strings.ToLower(title)
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	s = re.ReplaceAllString(s, "")
	reSpaces := regexp.MustCompile(`[\s-]+`)
	s = reSpaces.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

// ensureUniqueArticleSlug appends number suffix if slug already exists
func ensureUniqueArticleSlug(base string, excludeID uint) string {
	slug := base
	for i := 1; ; i++ {
		var count int64
		q := config.DB.Model(&models.Article{}).Where("slug = ?", slug)
		if excludeID > 0 {
			q = q.Where("id != ?", excludeID)
		}
		q.Count(&count)
		if count == 0 {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// GetArticles returns paginated list of articles
func GetArticles(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "12"))
	search := c.Query("search", "")
	tag := c.Query("tag", "")

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	query := config.DB.Model(&models.Article{}).Preload("User.Profile")

	if search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	var total int64
	query.Count(&total)

	var articles []models.Article
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&articles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch articles"})
	}

	return c.JSON(fiber.Map{
		"data":  articles,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetArticleBySlug returns a single article by slug
func GetArticleBySlug(c *fiber.Ctx) error {
	identifier := c.Params("slug")
	var article models.Article

	query := config.DB.Preload("User.Profile")
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&article).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Article retrieved successfully",
		"data":    article,
	})
}

// CreateArticle creates a new article
func CreateArticle(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if strings.TrimSpace(body.Title) == "" || strings.TrimSpace(body.Content) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title and content are required"})
	}

	baseSlug := generateArticleSlug(body.Title)
	slug := ensureUniqueArticleSlug(baseSlug, 0)

	article := models.Article{
		UserID:  uint(userID),
		Title:   body.Title,
		Content: body.Content,
		Tags:    body.Tags,
		Slug:    slug,
	}

	if err := config.DB.Create(&article).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create article"})
	}

	// Reload with user
	config.DB.Preload("User.Profile").First(&article, article.ID)

	return c.Status(201).JSON(fiber.Map{
		"message": "Article created successfully",
		"data":    article,
	})
}

// UpdateArticle updates an existing article
func UpdateArticle(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	identifier := c.Params("slug")

	var article models.Article
	query := config.DB
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&article).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
	}

	if article.UserID != uint(userID) {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if body.Title != "" {
		article.Title = body.Title
		newBase := generateArticleSlug(body.Title)
		article.Slug = ensureUniqueArticleSlug(newBase, article.ID)
	}
	if body.Content != "" {
		article.Content = body.Content
	}
	article.Tags = body.Tags

	if err := config.DB.Save(&article).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update article"})
	}

	return c.JSON(fiber.Map{
		"message": "Article updated successfully",
		"data":    article,
	})
}

// DeleteArticle deletes an article permanently
func DeleteArticle(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	identifier := c.Params("slug")

	var article models.Article
	query := config.DB
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&article).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
	}

	if article.UserID != uint(userID) {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Delete cover image if exists
	if article.CoverImage != "" {
		_ = os.Remove("." + article.CoverImage)
	}

	if err := config.DB.Unscoped().Delete(&article).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete article"})
	}

	return c.JSON(fiber.Map{"message": "Article deleted successfully"})
}

// UploadArticleCover uploads a cover image for an article
func UploadArticleCover(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	identifier := c.Params("slug")

	var article models.Article
	query := config.DB
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&article).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Article not found"})
	}

	if article.UserID != uint(userID) {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized"})
	}

	file, err := c.FormFile("cover")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Validate type
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return c.Status(400).JSON(fiber.Map{"error": "Only image files are allowed"})
	}

	// Remove old cover
	if article.CoverImage != "" {
		_ = os.Remove("." + article.CoverImage)
	}

	uploadDir := "./uploads/articles"
	_ = os.MkdirAll(uploadDir, 0755)

	ext := ".jpg"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "gif") {
		ext = ".gif"
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
	}

	filename := fmt.Sprintf("cover-%d-%d%s", article.ID, time.Now().UnixMilli(), ext)
	savePath := fmt.Sprintf("%s/%s", uploadDir, filename)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	}

	imageURL := fmt.Sprintf("/uploads/articles/%s", filename)
	config.DB.Model(&article).Update("cover_image", imageURL)

	return c.JSON(fiber.Map{
		"message":    "Cover uploaded successfully",
		"cover_image": imageURL,
	})
}
