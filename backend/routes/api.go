package routes

import (
	"kafapet-backend/config"
	"kafapet-backend/controllers"
	"kafapet-backend/middleware"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)
	auth.Get("/google", controllers.GoogleLogin)
	auth.Get("/google/callback", controllers.GoogleCallback)
	auth.Get("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")
		
		var user models.User
		if err := config.DB.Preload("Profile").First(&user, userID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		var totalLikes int64
		config.DB.Model(&models.Like{}).
			Joins("JOIN posts ON posts.id = likes.post_id").
			Where("posts.user_id = ?", userID).
			Count(&totalLikes)

		var followedIDs []uint
		config.DB.Model(&models.Connection{}).
			Where("follower_id = ?", userID).
			Pluck("followee_id", &followedIDs)

		return c.JSON(fiber.Map{
			"message":      "You are authenticated",
			"user_id":      userID,
			"role":         role,
			"profile":      user.Profile,
			"total_likes":  totalLikes,
			"followed_ids": followedIDs,
		})
	})

	// Alumni routes
	alumni := api.Group("/alumni")
	alumni.Get("/suggested", middleware.Protected(), controllers.GetSuggestedAlumni)
	alumni.Get("/", controllers.GetAlumniList) // Public access for MVP
	alumni.Put("/profile", middleware.Protected(), controllers.UpdateProfile)
	alumni.Put("/profile/avatar", middleware.Protected(), controllers.UpdateAvatar)
	alumni.Get("/:id", controllers.GetAlumniProfile)

	// Tags Routes
	api.Get("/tags/trending", controllers.GetTrendingTags)

	// Jobs Routes
	api.Get("/jobs", controllers.GetJobs)
	api.Post("/jobs", middleware.Protected(), controllers.CreateJob)
	api.Get("/jobs/:slug", controllers.GetJobBySlug)
	api.Put("/jobs/:slug", middleware.Protected(), controllers.UpdateJob)
	api.Delete("/jobs/:slug", middleware.Protected(), controllers.DeleteJob)
	api.Post("/jobs/:slug/banner", middleware.Protected(), controllers.UploadJobBanner)

	// Articles Routes
	api.Get("/articles", controllers.GetArticles)
	api.Post("/articles", middleware.Protected(), controllers.CreateArticle)
	api.Get("/articles/:slug", controllers.GetArticleBySlug)
	api.Put("/articles/:slug", middleware.Protected(), controllers.UpdateArticle)
	api.Delete("/articles/:slug", middleware.Protected(), controllers.DeleteArticle)
	api.Post("/articles/:slug/cover", middleware.Protected(), controllers.UploadArticleCover)

	// Posts (Social Network) Routes
	api.Get("/posts", controllers.GetPosts)
	api.Get("/posts/s/:slug", controllers.GetPostBySlug)
	api.Post("/posts", middleware.Protected(), controllers.CreatePost)
	api.Put("/posts/:id", middleware.Protected(), controllers.UpdatePost)
	api.Delete("/posts/:id", middleware.Protected(), controllers.DeletePost)

	// Like Routes
	api.Post("/posts/:id/like", middleware.Protected(), controllers.ToggleLike)
	api.Get("/posts/:id/likes", controllers.GetLikes)

	// Comment Routes
	api.Post("/posts/:id/comments", middleware.Protected(), controllers.CreateComment)
	api.Get("/posts/:id/comments", controllers.GetComments)
	api.Delete("/comments/:id", middleware.Protected(), controllers.DeleteComment)

	// Connection (Follow) Routes
	api.Post("/users/:id/follow", middleware.Protected(), controllers.FollowUser)
	api.Delete("/users/:id/follow", middleware.Protected(), controllers.UnfollowUser)
	api.Get("/users/:id/connections", controllers.GetUserConnections)
}
