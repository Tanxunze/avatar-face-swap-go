package handler

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

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

// ==================== Keycloak SSO Authentication ====================
// These endpoints match the original Python Flask implementation

// GET /api/login
// Redirects user to Keycloak for SSO authentication
func KeycloakLogin(c *gin.Context) {
	keycloak := service.GetKeycloakService()

	if !keycloak.IsEnabled() {
		response.Error(c, 400, "Keycloak SSO is not configured")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Build the callback URL (must match the registered redirect URI in Keycloak)
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/api/auth", scheme, c.Request.Host)

	authURL, err := keycloak.GetAuthorizationURL(ctx, redirectURI, "")
	if err != nil {
		service.LogActivity("ERROR", "Keycloak", "获取授权URL失败", "", "", c.ClientIP(), map[string]any{"error": err.Error()})
		response.Error(c, 500, "Failed to initiate Keycloak login")
		return
	}

	c.Redirect(302, authURL)
}

// GET /api/auth
// Keycloak callback handler - exchanges code for tokens and redirects to frontend
func KeycloakCallback(c *gin.Context) {
	keycloak := service.GetKeycloakService()
	cfg := config.Load()

	if !keycloak.IsEnabled() {
		response.Error(c, 400, "Keycloak SSO is not configured")
		return
	}

	// Get authorization code from query params
	code := c.Query("code")
	if code == "" {
		errorDesc := c.Query("error_description")
		if errorDesc == "" {
			errorDesc = c.Query("error")
		}
		service.LogActivity("ERROR", "Keycloak", "回调缺少授权码", "", "", c.ClientIP(), map[string]any{"error": errorDesc})
		c.Redirect(302, cfg.FrontendBaseURL+"/event?error="+url.QueryEscape("授权失败"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Build the same redirect URI used in the authorization request
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/api/auth", scheme, c.Request.Host)

	// Exchange code for tokens
	tokenResp, err := keycloak.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		service.LogActivity("ERROR", "Keycloak", "Token交换失败", "", "", c.ClientIP(), map[string]any{"error": err.Error()})
		c.Redirect(302, cfg.FrontendBaseURL+"/event?error="+url.QueryEscape("登录失败"))
		return
	}

	// Get user info from Keycloak
	userInfo, err := keycloak.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		service.LogActivity("ERROR", "Keycloak", "获取用户信息失败", "", "", c.ClientIP(), map[string]any{"error": err.Error()})
		c.Redirect(302, cfg.FrontendBaseURL+"/event?error="+url.QueryEscape("获取用户信息失败"))
		return
	}

	// SSO users are treated as admins (same as original Python implementation)
	role := "admin"
	username := userInfo.PreferredUsername
	if username == "" {
		username = userInfo.Sub
	}

	// Generate our JWT token
	jwtToken, err := service.GenerateJWT(username, role, userInfo.Email)
	if err != nil {
		service.LogActivity("ERROR", "Keycloak", "生成JWT失败", username, "", c.ClientIP(), map[string]any{"error": err.Error()})
		c.Redirect(302, cfg.FrontendBaseURL+"/event?error="+url.QueryEscape("生成令牌失败"))
		return
	}

	service.LogActivity("INFO", "用户认证", "SSO管理员登录", username, "", c.ClientIP(), map[string]any{
		"email":    userInfo.Email,
		"provider": "keycloak",
	})

	// Redirect to frontend with token (matches original: /event/admin?token=xxx)
	redirectURL := fmt.Sprintf("%s/event/%s?token=%s", cfg.FrontendBaseURL, role, jwtToken)
	c.Redirect(302, redirectURL)
}

// GET /api/logout
// Logout from Keycloak and redirect to frontend
func KeycloakLogout(c *gin.Context) {
	keycloak := service.GetKeycloakService()
	cfg := config.Load()

	if !keycloak.IsEnabled() {
		// If Keycloak is not configured, just redirect to frontend
		c.Redirect(302, cfg.FrontendBaseURL)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	logoutURL, err := keycloak.GetLogoutURL(ctx, cfg.FrontendBaseURL)
	if err != nil {
		// If we can't get logout URL, just redirect to frontend
		c.Redirect(302, cfg.FrontendBaseURL)
		return
	}

	c.Redirect(302, logoutURL)
}

// GET /api/profile
// Returns the current user's profile from session (Keycloak integration)
func GetProfile(c *gin.Context) {
	// In the Go implementation, we use JWT instead of session
	// This endpoint is called by frontend to check if user is logged in via SSO
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "Unauthorized")
		return
	}

	userEmail, _ := c.Get("user_email")
	role, _ := c.Get("role")

	response.Success(c, gin.H{
		"sub":                userID,
		"email":              userEmail,
		"preferred_username": userID,
		"role":               role,
	})
}
