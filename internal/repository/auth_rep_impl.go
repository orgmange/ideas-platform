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

func (r *authRepository) GetUserByPhone(phone string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("user", phone)
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) CreateUser(user *models.User) (*models.User, error) {
	err := r.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *authRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("role", name)
		}
		return nil, err
	}
	return &role, nil
}

// CreateRefreshToken creates a new refresh token.
func (r *authRepository) CreateRefreshToken(token *models.UserRefreshToken) error {
	return r.db.Create(token).Error
}

// GetRefreshToken retrieves a token by its value.
func (r *authRepository) GetRefreshToken(token string) (*models.UserRefreshToken, error) {
	var refreshToken models.UserRefreshToken
	if err := r.db.Preload("User.Role").First(&refreshToken, "refresh_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("refresh token", token)
		}
		return nil, err
	}
	return &refreshToken, nil
}

// DeleteRefreshToken deletes a token by its value.
func (r *authRepository) DeleteRefreshToken(token string) error {
	result := r.db.Where("refresh_token = ?", token).Delete(&models.UserRefreshToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("refresh token", token)
	}
	return nil
}

// DeleteRefreshTokensByUserID deletes all refresh tokens for a specific user.
func (r *authRepository) DeleteRefreshTokensByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.UserRefreshToken{}).Error
}
