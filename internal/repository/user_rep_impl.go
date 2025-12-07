package repository

import (
	"context"
	"fmt"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRep {
	return &UserRepImpl{db: db}
}

func (u *UserRepImpl) DeleteUser(ctx context.Context, ID uuid.UUID) error {
	result := u.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", ID).Update("is_deleted", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("user", ID.String())
	}
	return nil
}

func (u *UserRepImpl) GetAllUsers(ctx context.Context, limit int, offset int) ([]models.User, error) {
	var users []models.User
	err := u.db.WithContext(ctx).Where("is_deleted = ?", false).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (u *UserRepImpl) GetUser(ctx context.Context, ID uuid.UUID) (*models.User, error) {
	var user models.User
	err := u.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", ID, false).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("user", ID.String())
		}
		return nil, err
	}
	return &user, nil
}

func (u *UserRepImpl) UpdateUser(ctx context.Context, user *models.User) error {
	return u.db.WithContext(ctx).Save(user).Error
}

func (u *UserRepImpl) IsUserExist(ctx context.Context, ID uuid.UUID) (bool, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", ID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}
