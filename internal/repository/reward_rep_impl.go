package repository

import (
	"context"
	"errors"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardRepositoryImpl struct {
	db *gorm.DB
}

func NewRewardRepository(db *gorm.DB) RewardRepository {
	return &RewardRepositoryImpl{db: db}
}

func (r *RewardRepositoryImpl) CreateReward(ctx context.Context, reward *models.Reward) (*models.Reward, error) {
	if err := r.db.WithContext(ctx).Create(reward).Error; err != nil {
		return nil, err
	}
	return reward, nil
}

func (r *RewardRepositoryImpl) UpdateReward(ctx context.Context, reward *models.Reward) error {
	return r.db.WithContext(ctx).Save(reward).Error
}

func (r *RewardRepositoryImpl) DeleteReward(ctx context.Context, ID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Reward{}, ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("reward", ID.String())
	}
	return nil
}

func (r *RewardRepositoryImpl) GetReward(ctx context.Context, ID uuid.UUID) (*models.Reward, error) {
	var reward models.Reward
	if err := r.db.WithContext(ctx).First(&reward, "id = ?", ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewErrNotFound("reward", ID.String())
		}
		return nil, err
	}
	return &reward, nil
}

func (r *RewardRepositoryImpl) GetRewardsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]models.Reward, error) {
	var rewards []models.Reward
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Where("receiver_id = ?", userID).Find(&rewards).Error; err != nil {
		return nil, err
	}
	return rewards, nil
}

func (r *RewardRepositoryImpl) GetRewardsByCoffeeShopID(ctx context.Context, coffeeShopID uuid.UUID, offset, limit int) ([]models.Reward, error) {
	var rewards []models.Reward
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Where("coffee_shop_id = ?", coffeeShopID).Find(&rewards).Error; err != nil {
		return nil, err
	}
	return rewards, nil
}
