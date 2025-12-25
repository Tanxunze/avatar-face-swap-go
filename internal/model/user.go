package model

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserID    string `json:"sub"`
	Role      string `json:"role"`
	UserEmail string `json:"user_email,omitempty"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Token string `json:"token" binding:"required"`
}

type LoginResponse struct {
	EventID     string `json:"event_id"`
	Description string `json:"description,omitempty"`
	Token       string `json:"token"`
}
