package repository

import (
	"context"

	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type AuthRepository interface {
	GetOTP(ctx context.Context, phone string) (*models.OTP, error)
	CreateOTP(ctx context.Context, otp *models.OTP) error
	UpdateOTP(ctx context.Context, otp *models.OTP) error
	DeleteOTP(ctx context.Context, phone string) error
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)

	// Refresh Token
	CreateRefreshToken(ctx context.Context, token *models.UserRefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*models.UserRefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokensByUserID(ctx context.Context, userID uuid.UUID) error
}
