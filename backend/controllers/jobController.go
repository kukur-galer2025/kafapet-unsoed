package controllers

import (
	"fmt"
	"kafapet-backend/config"
	"kafapet-backend/models"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// generateJobSlug creates a URL-safe slug from title and company
func generateJobSlug(title, company string) string {
	combined := strings.ToLower(title + " " + company)
	// Replace spaces and special chars with hyphens
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	combined = re.ReplaceAllString(combined, "")
	re2 := regexp.MustCompile(`[\s-]+`)
	combined = re2.ReplaceAllString(combined, "-")
	combined = strings.Trim(combined, "-")
	if len(combined) > 80 {
		combined = combined[:80]
	}
	return combined
}

// ensureUniqueJobSlug makes slug unique by appending a number if needed
func ensureUniqueJobSlug(base string, excludeID uint) string {
	slug := base
	for i := 1; ; i++ {
		var count int64
		query := config.DB.Model(&models.Job{}).Where("slug = ?", slug)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		query.Count(&count)
		if count == 0 {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// GetJobs returns a list of jobs, optionally filtered
func GetJobs(c *fiber.Ctx) error {
	var jobs []models.Job

	query := config.DB.Preload("User.Profile").Order("created_at desc")

	search := c.Query("search")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title LIKE ? OR company LIKE ? OR location LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	jobType := c.Query("type")
	if jobType != "" {
		query = query.Where("job_type = ?", jobType)
	}

	if err := query.Find(&jobs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch jobs"})
	}

	return c.JSON(fiber.Map{
		"message": "Jobs retrieved successfully",
		"data":    jobs,
	})
}

// GetJobBySlug returns a single job detail by slug or ID
func GetJobBySlug(c *fiber.Ctx) error {
	identifier := c.Params("slug")
	var job models.Job

	query := config.DB.Preload("User.Profile")
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&job).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Job retrieved successfully",
		"data":    job,
	})
}

// CreateJob creates a new job posting
func CreateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)

	job := new(models.Job)
	if err := c.BodyParser(job); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if job.Title == "" || job.Company == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title and Company are required"})
	}

	job.UserID = uint(userID)
	job.Status = "open"
	baseSlug := generateJobSlug(job.Title, job.Company)
	job.Slug = ensureUniqueJobSlug(baseSlug, 0)

	if err := config.DB.Create(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create job"})
	}

	config.DB.Preload("User.Profile").First(&job, job.ID)

	return c.Status(201).JSON(fiber.Map{
		"message": "Job created successfully",
		"data":    job,
	})
}

// UpdateJob updates a job posting
func UpdateJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	identifier := c.Params("slug")

	var job models.Job
	query := config.DB
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&job).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}

	if job.UserID != uint(userID) {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized to update this job"})
	}

	var updateData models.Job
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	job.Title = updateData.Title
	job.Company = updateData.Company
	job.Location = updateData.Location
	job.JobType = updateData.JobType
	job.SalaryRange = updateData.SalaryRange
	job.ApplyLink = updateData.ApplyLink
	job.Description = updateData.Description
	job.Requirements = updateData.Requirements

	// Regenerate slug if title or company changed
	newBase := generateJobSlug(job.Title, job.Company)
	job.Slug = ensureUniqueJobSlug(newBase, job.ID)

	if err := config.DB.Save(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update job"})
	}

	config.DB.Preload("User.Profile").First(&job, job.ID)

	return c.JSON(fiber.Map{
		"message": "Job updated successfully",
		"data":    job,
	})
}

// DeleteJob deletes a job if requested by the owner
func DeleteJob(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	identifier := c.Params("slug")

	var job models.Job
	query := config.DB
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&job).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}

	if job.UserID != uint(userID) {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized to delete this job"})
	}

	if err := config.DB.Unscoped().Delete(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete job"})
	}

	return c.JSON(fiber.Map{
		"message": "Job deleted successfully",
	})
}

// UploadJobBanner uploads a banner image for a specific job
func UploadJobBanner(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	identifier := c.Params("slug")

	var job models.Job
	query := config.DB
	if id, err := strconv.Atoi(identifier); err == nil {
		query = query.Where("id = ? OR slug = ?", id, identifier)
	} else {
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&job).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}

	if job.UserID != uint(userID) {
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized"})
	}

	file, err := c.FormFile("banner")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to get banner file"})
	}

	uploadDir := "./uploads/jobs/banners"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create directory"})
	}

	filePath := uploadDir + "/" + file.Filename
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save banner"})
	}

	job.BannerImage = "http://localhost:3000/uploads/jobs/banners/" + file.Filename
	if err := config.DB.Save(&job).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update job banner"})
	}

	return c.JSON(fiber.Map{
		"message":      "Banner uploaded successfully",
		"banner_image": job.BannerImage,
	})
}
