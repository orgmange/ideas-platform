package repository

import (
	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) GetOTP(phone string) (*models.OTP, error) {
	var otp models.OTP
	if err := r.db.Where("phone = ? AND verified = ?", phone, false).First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("user", phone)
		}
		return nil, err
	}
	return &otp, nil
}

func (r *authRepository) CreateOTP(otp *models.OTP) error {
	return r.db.Create(otp).Error
}

func (r *authRepository) UpdateOTP(otp *models.OTP) error {
	return r.db.Save(otp).Error
}

func (r *authRepository) DeleteOTP(phone string) error {
	return r.db.Where("phone = ?", phone).Delete(&models.OTP{}).Error
}

func (r *authRepository) GetUserIDByPhone(phone string) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.Model(&models.User{}).Where("phone = ?", phone).Pluck("id", &id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("user", phone)
		}
		return nil, err
	}
	return &id, nil
}

func (r *authRepository) CreateUser(user *models.User) (*uuid.UUID, error) {
	err := r.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return &user.ID, nil
}
