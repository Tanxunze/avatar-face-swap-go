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

	if err := database.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer database.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	api := router.Group("/api")
	{
		// Auth routes
		api.POST("/verify", handler.Login)
		api.POST("/verify-token", handler.VerifyToken)

		// Event routes
		api.GET("/events", middleware.AuthRequired(), middleware.AdminRequired(), handler.ListEvents)
		api.GET("/events/:id", middleware.AuthRequired(), handler.GetEvent)
		api.POST("/events", middleware.AuthRequired(), middleware.AdminRequired(), handler.CreateEvent)
		api.PUT("/events/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.UpdateEvent)
		api.DELETE("/events/:id", middleware.AuthRequired(), middleware.AdminRequired(), handler.DeleteEvent)
	}

	log.Printf("Server starting on :%s", cfg.Port)
	router.Run(":" + cfg.Port)
}
