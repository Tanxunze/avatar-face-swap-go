package main

import (
	"log"

	"avatar-face-swap-go/internal/config"
	"avatar-face-swap-go/internal/database"
	"avatar-face-swap-go/internal/handler"
	"avatar-face-swap-go/internal/middleware"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := database.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer database.Close()

	router := gin.Default()

	// CORS configuration from environment
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.GetCORSOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	api := router.Group("/api")
	{
		// Auth - Token based (local admin password / event token)
		auth := api.Group("/auth")
		{
			auth.POST("/sessions", handler.Login)           // Create session (login)
			auth.POST("/tokens/verify", handler.VerifyToken) // Verify JWT token

			// SSO (Keycloak)
			auth.GET("/sso/login", handler.KeycloakLogin)       // Redirect to Keycloak
			auth.GET("/sso/callback", handler.KeycloakCallback) // Keycloak callback
			auth.DELETE("/sessions/current", handler.KeycloakLogout) // Logout
			auth.GET("/profile", middleware.AuthRequired(), handler.GetProfile) // Get user profile
		}

		// Event
		api.GET("/events", middleware.AuthRequired(), middleware.AdminRequired(), handler.ListEvents)
		api.GET("/events/:id", middleware.AuthRequired(), handler.GetEvent)
		api.POST("/events", middleware.AuthRequired(), middleware.AdminRequired(), handler.CreateEvent)
		api.PUT("/events/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.UpdateEvent)
		api.DELETE("/events/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.DeleteEvent)
		api.GET("/events/:id/token", middleware.AuthRequired(), middleware.AdminRequired(), handler.GetEventToken)
		api.GET("/events/:id/status", middleware.AuthRequired(), middleware.AdminRequired(), handler.GetProcessStatus) // Get face detection status

		// Picture (event main image)
		api.GET("/events/:id/picture", middleware.AuthRequired(), handler.GetEventPic)                                       // Get event picture
		api.GET("/events/:id/picture/metadata", middleware.AuthRequired(), middleware.AdminRequired(), handler.GetEventPicInfo) // Get picture metadata
		api.PUT("/events/:id/picture", middleware.AuthRequired(), middleware.AdminRequired(), handler.UploadEventPic)        // Upload/replace event picture

		// Faces
		api.GET("/events/:id/faces", middleware.AuthRequired(), handler.GetEventFaces)                                           // List faces
		api.GET("/events/:id/faces/metadata", middleware.AuthRequired(), middleware.AdminRequired(), handler.GetEventMetadata)   // Get faces metadata
		api.GET("/events/:id/faces/:filename", middleware.AuthRequired(), handler.GetFaceImage)                                  // Get face image
		api.POST("/events/:id/faces", middleware.AuthRequired(), middleware.AdminRequired(), handler.AddManualFace)              // Add manual face
		api.POST("/events/:id/faces/:face/avatar", middleware.AuthRequired(), handler.UploadAvatar)       // Upload avatar for a face
		api.GET("/events/:id/avatars/:filename", middleware.AuthRequired(), handler.GetUploadedAvatar)    // Get uploaded avatar
		api.DELETE("/events/:id/faces/:filename", middleware.AuthRequired(), middleware.AdminRequired(), handler.DeleteFace)

		// QQ integration
		api.GET("/events/:id/qq-profiles/:qq", middleware.AuthRequired(), handler.GetQQNickname)              // Get QQ nickname
		api.POST("/events/:id/faces/:face/qq-avatar", middleware.AuthRequired(), handler.UploadQQAvatar)      // Upload QQ avatar for a face
		api.GET("/events/:id/faces/:filename/qq-profile", middleware.AuthRequired(), handler.GetFaceQQInfo)   // Get QQ info for a face

		// log
		api.GET("/logs", middleware.AuthRequired(), middleware.AdminRequired(), handler.GetLogs)
	}

	log.Printf("Server starting on :%s", cfg.Port)
	router.Run(":" + cfg.Port)
}
