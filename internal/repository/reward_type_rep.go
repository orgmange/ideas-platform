package repository

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type RewardTypeRepository interface {
	GetRewardType(ctx context.Context, rewardTypeID uuid.UUID) (*models.RewardType, error)
	GetRewardsTypeByCoffeeShopID(ctx context.Context, CoffeeShopID uuid.UUID, offset, limit int) ([]models.RewardType, error)
	UpdateReward(ctx context.Context, rewardType *models.RewardType) error
	DeleteReward(ctx context.Context, rewardTypeID uuid.UUID) error
	CreateReward(ctx context.Context, rewardType *models.RewardType) (*models.RewardType, error)
}
