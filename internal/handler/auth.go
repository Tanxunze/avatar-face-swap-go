package handler

import (
	"strconv"

	"avatar-face-swap-go/internal/config"
	"avatar-face-swap-go/internal/model"
	"avatar-face-swap-go/internal/repository"
	"avatar-face-swap-go/internal/service"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// POST /api/verify
func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Missing token")
		return
	}

	cfg := config.Load()

	// Check if admin password
	if req.Token == cfg.AdminPassword {
		jwtToken, err := service.GenerateJWT("local_admin", "admin", "")
		if err != nil {
			response.Error(c, 500, "Failed to generate token")
			return
		}

		service.LogActivity("INFO", "用户认证", "管理员登录", "local_admin", "", c.ClientIP(), nil)

		response.Success(c, model.LoginResponse{
			EventID: "admin",
			Token:   jwtToken,
		})
		return
	}

	// Check if event token
	event, err := repository.GetEventByToken(req.Token)
	if err != nil {
		response.Error(c, 500, "Database error")
		return
	}

	if event == nil {
		response.Error(c, 404, "Invalid token")
		return
	}

	if !event.IsOpen {
		response.Error(c, 400, "Event is not open")
		return
	}

	// Generate JWT with event_id as role
	jwtToken, err := service.GenerateJWT("local_user", formatEventID(event.ID), "")
	if err != nil {
		response.Error(c, 500, "Failed to generate token")
		return
	}

	service.LogActivity("INFO", "用户认证", "用户登录", "", formatEventID(event.ID), c.ClientIP(), nil)

	response.Success(c, model.LoginResponse{
		EventID:     formatEventID(event.ID),
		Description: event.Description,
		Token:       jwtToken,
	})
}

// POST /api/verify-token
func VerifyToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Missing token")
		return
	}

	claims, err := service.ValidateJWT(req.Token)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	response.Success(c, gin.H{
		"user":     claims.UserID,
		"role":     claims.Role,
		"event_id": claims.Role,
	})
}

func formatEventID(id int) string {
	return strconv.Itoa(id)
}
