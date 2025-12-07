package repository

import (
	"context"
	"errors"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardTypeRepositoryImpl struct {
	db *gorm.DB
}

func NewRewardTypeRepository(db *gorm.DB) RewardTypeRepository {
	return RewardTypeRepositoryImpl{
		db: db,
	}
}

func (r RewardTypeRepositoryImpl) CreateReward(ctx context.Context, rewardType *models.RewardType) (*models.RewardType, error) {
	err := r.db.WithContext(ctx).Create(rewardType).Error
	if err != nil {
		return nil, err
	}

	return rewardType, nil
}

func (r RewardTypeRepositoryImpl) DeleteReward(ctx context.Context, rewardTypeID uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.RewardType{}, rewardTypeID)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return apperrors.NewErrNotFound("rewardType", rewardTypeID.String())
	}

	return nil
}

func (r RewardTypeRepositoryImpl) GetRewardType(ctx context.Context, rewardTypeID uuid.UUID) (*models.RewardType, error) {
	var rewardType models.RewardType
	err := r.db.WithContext(ctx).Take(&rewardType, rewardTypeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewErrNotFound("rewardType", rewardTypeID.String())
		}
		return nil, err
	}

	return &rewardType, nil
}

func (r RewardTypeRepositoryImpl) GetRewardsTypeByCoffeeShopID(ctx context.Context, CoffeeShopID uuid.UUID, offset int, limit int) ([]models.RewardType, error) {
	var rewardTypes []models.RewardType
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Where("coffee_shop_id = ?", CoffeeShopID).Find(&rewardTypes).Error
	if err != nil {
		return rewardTypes, err
	}

	return rewardTypes, nil
}

func (r RewardTypeRepositoryImpl) UpdateReward(ctx context.Context, rewardType *models.RewardType) error {
	return r.db.WithContext(ctx).Save(rewardType).Error
}
