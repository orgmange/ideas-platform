package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type AuthUsecase interface {
	GetOTP(phone string) error
	VerifyOTP(req *dto.VerifyOTPRequest) (*dto.AuthResponse, error)
	Refresh(token string) (*dto.AuthResponse, error)
	Logout(token string) error
	LogoutEverywhere(userID uuid.UUID) error
	ValidateJWTToken(tokenString string) (*dto.JWTClaims, error)
}
