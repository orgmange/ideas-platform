package repository

import (
	"context"
	"fmt" // Added this line

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

func (r *authRepository) GetOTP(ctx context.Context, phone string) (*models.OTP, error) {
	var otp models.OTP
	if err := r.db.WithContext(ctx).Where("phone = ? AND verified = ?", phone, false).First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("user", phone)
		}
		return nil, err
	}
	return &otp, nil
}

func (r *authRepository) CreateOTP(ctx context.Context, otp *models.OTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *authRepository) UpdateOTP(ctx context.Context, otp *models.OTP) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

func (r *authRepository) DeleteOTP(ctx context.Context, phone string) error {
	return r.db.WithContext(ctx).Where("phone = ?", phone).Delete(&models.OTP{}).Error
}

func (r *authRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("user", phone)
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *authRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("role", name)
		}
		return nil, err
	}
	return &role, nil
}

// CreateRefreshToken creates a new refresh token.
func (r *authRepository) CreateRefreshToken(ctx context.Context, token *models.UserRefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetRefreshToken retrieves a token by its value.
func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (*models.UserRefreshToken, error) {
	var refreshToken models.UserRefreshToken
	if err := r.db.WithContext(ctx).Preload("User").First(&refreshToken, "refresh_token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("refresh token", token)
		}
		return nil, err
	}
	return &refreshToken, nil
}

// DeleteRefreshToken deletes a token by its value.
func (r *authRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Where("refresh_token = ?", token).Delete(&models.UserRefreshToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("refresh token", token)
	}
	return nil
}

// DeleteRefreshTokensByUserID deletes all refresh tokens for a specific user.
func (r *authRepository) DeleteRefreshTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserRefreshToken{})
	if result.Error != nil {
		fmt.Printf("DeleteRefreshTokensByUserID failed for userID %s: %v\n", userID, result.Error)
		return result.Error
	}
	fmt.Printf("DeleteRefreshTokensByUserID affected %d rows for userID %s\n", result.RowsAffected, userID)
	return nil
}
