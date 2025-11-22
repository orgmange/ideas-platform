package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type RewardTypeRepository interface {
	GetRewardType(rewardTypeID uuid.UUID) (*models.RewardType, error)
	GetRewardsTypeByCoffeeShopID(CoffeeShopID uuid.UUID, offset, limit int) ([]models.RewardType, error)
	UpdateReward(rewardType *models.RewardType) error
	DeleteReward(rewardTypeID uuid.UUID) error
	CreateReward(rewardType *models.RewardType) (*models.RewardType, error)
}
