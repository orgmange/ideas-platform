package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type AuthRepository interface {
	GetOTP(phone string) (*models.OTP, error)
	CreateOTP(otp *models.OTP) error
	UpdateOTP(otp *models.OTP) error
	DeleteOTP(phone string) error
	GetUserByPhone(phone string) (*models.User, error)
	CreateUser(*models.User) (*models.User, error)
	GetRoleByName(name string) (*models.Role, error)

	// Refresh Token
	CreateRefreshToken(token *models.UserRefreshToken) error
	GetRefreshToken(token string) (*models.UserRefreshToken, error)
	DeleteRefreshToken(token string) error
	DeleteRefreshTokensByUserID(userID uuid.UUID) error
}
