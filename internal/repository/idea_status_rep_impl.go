package repository

import (
	"context"

	"github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IdeaStatusRepositoryImpl struct {
	db *gorm.DB
}

func NewIdeaStatusRepository(db *gorm.DB) IdeaStatusRepository {
	return &IdeaStatusRepositoryImpl{db: db}
}

func (r *IdeaStatusRepositoryImpl) Create(ctx context.Context, status *models.IdeaStatus) (uuid.UUID, error) {
	if err := r.db.WithContext(ctx).Create(status).Error; err != nil {
		return uuid.Nil, err
	}
	return status.ID, nil
}

func (r *IdeaStatusRepositoryImpl) Update(ctx context.Context, status *models.IdeaStatus) error {
	return r.db.WithContext(ctx).Save(status).Error
}

func (r *IdeaStatusRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.IdeaStatus{}).Where("id = ?", id).Update("is_deleted", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("idea status", id.String())
	}
	return nil
}

func (r *IdeaStatusRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (models.IdeaStatus, error) {
	var status models.IdeaStatus
	err := r.db.WithContext(ctx).First(&status, "id = ? AND is_deleted = false", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return status, apperrors.NewErrNotFound("idea status", id.String())
		}
		return status, err
	}
	return status, nil
}

func (r *IdeaStatusRepositoryImpl) GetAll(ctx context.Context) ([]models.IdeaStatus, error) {
	var statuses []models.IdeaStatus
	err := r.db.WithContext(ctx).Where("is_deleted = false").Find(&statuses).Error
	return statuses, err
}

func (r *IdeaStatusRepositoryImpl) GetByTitle(ctx context.Context, title string) (models.IdeaStatus, error) {
	var status models.IdeaStatus
	err := r.db.WithContext(ctx).First(&status, "title = ? AND is_deleted = false", title).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return status, apperrors.NewErrNotFound("idea status", title)
		}
		return status, err
	}
	return status, err
}