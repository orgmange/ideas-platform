package usecase

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type RewardUsecase interface {
	GiveReward(ctx context.Context, actorID uuid.UUID, req *dto.GiveRewardRequest) (*dto.RewardResponse, error)
	RevokeReward(ctx context.Context, actorID, rewardID uuid.UUID) error
	GetReward(ctx context.Context, rewardID uuid.UUID) (*dto.RewardResponse, error)
	GetRewardsForCoffeeShop(ctx context.Context, actorID, coffeeShopID uuid.UUID, page, limit int) ([]dto.RewardResponse, error)
	GetMyRewards(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.RewardResponse, error)
}
