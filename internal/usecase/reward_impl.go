package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type RewardUsecaseImpl struct {
	rewardRepo repository.RewardRepository
	ideaRepo   repository.IdeaRepository
	logger     *slog.Logger
}

func NewRewardUsecase(rewardRepo repository.RewardRepository, ideaRepo repository.IdeaRepository, logger *slog.Logger) RewardUsecase {
	return &RewardUsecaseImpl{
		rewardRepo: rewardRepo,
		ideaRepo:   ideaRepo,
		logger:     logger,
	}
}

func (u *RewardUsecaseImpl) GiveReward(ctx context.Context, actorID uuid.UUID, req *dto.GiveRewardRequest) (*dto.RewardResponse, error) {
	logger := u.logger.With("method", "GiveReward", "adminID", actorID.String(), "ideaID", req.IdeaID.String())
	logger.Debug("starting to give a reward")

	idea, err := u.ideaRepo.GetIdea(ctx, req.IdeaID)
	if err != nil {
		logger.Error("failed to get idea", "error", err)
		return nil, err
	}

	now := time.Now()
	reward := &models.Reward{
		ReceiverID:   idea.CreatorID,
		CoffeeShopID: idea.CoffeeShopID,
		IdeaID:       &idea.ID,
		RewardTypeID: &req.RewardTypeID,
		IsActivated:  false,
		GivenAt:      &now,
	}

	createdReward, err := u.rewardRepo.CreateReward(ctx, reward)
	if err != nil {
		logger.Error("failed to create reward in repository", "error", err)
		return nil, err
	}

	logger.Info("reward given successfully by admin", "rewardID", createdReward.ID)
	return toRewardResponse(createdReward), nil
}

func (u *RewardUsecaseImpl) RevokeReward(ctx context.Context, actorID, rewardID uuid.UUID) error {
	logger := u.logger.With("method", "RevokeReward", "adminID", actorID.String(), "rewardID", rewardID.String())
	logger.Debug("starting to revoke a reward by admin")

	// Check if reward exists before deleting
	if _, err := u.rewardRepo.GetReward(ctx, rewardID); err != nil {
		logger.Error("failed to get reward for deletion", "error", err)
		return err
	}

	if err := u.rewardRepo.DeleteReward(ctx, rewardID); err != nil {
		logger.Error("failed to delete reward from repository", "error", err)
		return err
	}

	logger.Info("reward revoked successfully by admin")
	return nil
}

func (u *RewardUsecaseImpl) GetReward(ctx context.Context, rewardID uuid.UUID) (*dto.RewardResponse, error) {
	logger := u.logger.With("method", "GetReward", "rewardID", rewardID.String())
	logger.Debug("starting to get a reward")

	reward, err := u.rewardRepo.GetReward(ctx, rewardID)
	if err != nil {
		logger.Error("failed to get reward", "error", err)
		return nil, err
	}

	logger.Info("reward fetched successfully")
	return toRewardResponse(reward), nil
}

func (u *RewardUsecaseImpl) GetRewardsForCoffeeShop(ctx context.Context, actorID, coffeeShopID uuid.UUID, page, limit int) ([]dto.RewardResponse, error) {
	logger := u.logger.With("method", "GetRewardsForCoffeeShop", "actorID", actorID.String(), "coffeeShopID", coffeeShopID.String())
	logger.Debug("starting to get rewards for coffee shop")

	if limit <= 0 || limit > 50 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}

	rewards, err := u.rewardRepo.GetRewardsByCoffeeShopID(ctx, coffeeShopID, page*limit, limit)
	if err != nil {
		logger.Error("failed to get rewards for coffee shop", "error", err)
		return nil, err
	}

	logger.Info("rewards for coffee shop fetched successfully", "count", len(rewards))
	return toRewardResponses(rewards), nil
}

func (u *RewardUsecaseImpl) GetMyRewards(ctx context.Context, userID uuid.UUID, page, limit int) ([]dto.RewardResponse, error) {
	logger := u.logger.With("method", "GetMyRewards", "userID", userID.String())
	logger.Debug("starting to get my rewards")

	if limit <= 0 || limit > 50 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}

	rewards, err := u.rewardRepo.GetRewardsByUserID(ctx, userID, page*limit, limit)
	if err != nil {
		logger.Error("failed to get rewards for user", "error", err)
		return nil, err
	}

	logger.Info("user rewards fetched successfully", "count", len(rewards))
	return toRewardResponses(rewards), nil
}

func toRewardResponse(r *models.Reward) *dto.RewardResponse {
	return &dto.RewardResponse{
		ID:           r.ID,
		ReceiverID:   r.ReceiverID,
		CoffeeShopID: r.CoffeeShopID,
		IdeaID:       r.IdeaID,
		RewardTypeID: r.RewardTypeID,
		IsActivated:  r.IsActivated,
		GivenAt:      r.GivenAt,
		CreatedAt:    r.CreatedAt,
	}
}

func toRewardResponses(rewards []models.Reward) []dto.RewardResponse {
	res := make([]dto.RewardResponse, len(rewards))
	for i, r := range rewards {
		res[i] = *toRewardResponse(&r)
	}
	return res
}
