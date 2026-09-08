package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"kafapet-backend/config"
	"kafapet-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig *oauth2.Config
)

// InitOauthConfig is called to initialize config from env
func InitOauthConfig() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleLogin redirects user to Google Consent page
func GoogleLogin(c *fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL("random-state-string")
	return c.Redirect(url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the response from Google
func GoogleCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	if state != "random-state-string" {
		return c.SendString("Invalid state")
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.SendString("Code exchange failed: " + err.Error())
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return c.SendString("Failed to get user info: " + err.Error())
	}
	defer resp.Body.Close()

	userData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.SendString("Failed to read user info: " + err.Error())
	}

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(userData, &googleUser); err != nil {
		return c.SendString("Failed to parse user info")
	}

	var user models.User
	result := config.DB.Where("email = ?", googleUser.Email).First(&user)

	// If user doesn't exist, create one
	if result.Error != nil {
		user = models.User{
			Email:      googleUser.Email,
			PasswordHash: "", // No password for Google auth
			Provider:   "google",
			ProviderID: googleUser.ID,
			Role:       "alumni",
		}
		config.DB.Create(&user)
		
		// Create empty profile with data from google
		profile := models.Profile{
			UserID:     user.ID,
			FullName:   googleUser.Name,
			FotoProfil: googleUser.Picture,
		}
		config.DB.Create(&profile)
	}

	// Generate JWT Token
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret"
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// Redirect back to frontend with the token in URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4321"
	}

	return c.Redirect(frontendURL+"/auth/callback?token="+t, http.StatusFound)
}
