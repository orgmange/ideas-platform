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
	GetUserIDByPhone(phone string) (*uuid.UUID, error)
	CreateUser(*models.User) (*uuid.UUID, error)
}
