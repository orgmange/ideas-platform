package repository

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type RewardRepository interface {
	GetReward(ctx context.Context, rewardID uuid.UUID) (*models.Reward, error)
	GetRewardsByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]models.Reward, error)
	GetRewardsByCoffeeShopID(ctx context.Context, CoffeeShopID uuid.UUID, offset, limit int) ([]models.Reward, error)
	UpdateReward(ctx context.Context, reward *models.Reward) error
	DeleteReward(ctx context.Context, rewardID uuid.UUID) error
	CreateReward(ctx context.Context, reward *models.Reward) (*models.Reward, error)
}
