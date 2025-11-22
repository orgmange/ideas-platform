package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type RewardRepository interface {
	GetReward(rewardID uuid.UUID) (*models.Reward, error)
	GetRewardsByUserID(userID uuid.UUID, offset, limit int) ([]models.Reward, error)
	GetRewardsByCoffeeShopID(CoffeeShopID uuid.UUID, offset, limit int) ([]models.Reward, error)
	UpdateReward(reward *models.Reward) error
	DeleteReward(rewardID uuid.UUID) error
	CreateReward(reward *models.Reward) (*models.Reward, error)
}
