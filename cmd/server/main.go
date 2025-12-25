package main

import (
	"log"

	"avatar-face-swap-go/internal/config"
	"avatar-face-swap-go/internal/database"
	"avatar-face-swap-go/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer database.Close()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	api := router.Group("/api")
	{
		api.GET("/events", handler.ListEvents)
		api.GET("/events/:id", handler.GetEvent)
		api.POST("/events", handler.CreateEvent)
	}

	log.Printf("Server starting on :%s", cfg.Port)
	router.Run(":" + cfg.Port)
}
