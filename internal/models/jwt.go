package models

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID       string `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}
