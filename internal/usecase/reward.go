package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type RewardUsecase interface {
	GiveReward(actorID uuid.UUID, req *dto.GiveRewardRequest) (*dto.RewardResponse, error)
	RevokeReward(actorID, rewardID uuid.UUID) error
	GetReward(rewardID uuid.UUID) (*dto.RewardResponse, error)
	GetRewardsForCoffeeShop(actorID, coffeeShopID uuid.UUID, page, limit int) ([]dto.RewardResponse, error)
	GetMyRewards(userID uuid.UUID, page, limit int) ([]dto.RewardResponse, error)
}

