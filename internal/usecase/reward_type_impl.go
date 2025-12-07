package usecase

import (
	"context"
	"log/slog"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type RewardTypeUsecaseImpl struct {
	rep         repository.RewardTypeRepository
	csRep       repository.CoffeeShopRep
	workerCsRep repository.WorkerCoffeeShopRepository
	logger      *slog.Logger
}

func NewRewardTypeUsecase(rep repository.RewardTypeRepository,
	csRep repository.CoffeeShopRep,
	workerCsRep repository.WorkerCoffeeShopRepository,
	logger *slog.Logger,
) RewardTypeUsecase {
	return &RewardTypeUsecaseImpl{
		rep:         rep,
		csRep:       csRep,
		workerCsRep: workerCsRep,
		logger:      logger,
	}
}

// CreateRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) CreateRewardType(ctx context.Context, creatorID uuid.UUID, request *dto.CreateRewardTypeRequest) (*dto.RewardTypeResponse, error) {
	logger := r.logger.With("method", "CreateRewardType", "creator id", creatorID.String(), "coffee shop id", request.CoffeeShopID.String())
	logger.Debug("starting create reward type")

	err := CheckShopAdminAccess(ctx, logger, r.workerCsRep, creatorID, request.CoffeeShopID)
	if err != nil {
		return nil, err
	}

	rewardType := toRewardType(request)
	savedRewardType, err := r.rep.CreateReward(ctx, rewardType)
	if err != nil {
		logger.Error("unexpected error when creating reward type", "error", err.Error())
		return nil, err
	}

	logger.Error("reward type created successfully", "reward type id", savedRewardType.ID.String())
	return toRewardTypeResponse(savedRewardType), nil
}

// DeleteRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) DeleteRewardType(ctx context.Context, deleterID uuid.UUID, rewardTypeID uuid.UUID) error {
	logger := r.logger.With("method", "DeleteRewardType", "deleter id", deleterID.String())
	logger.Debug("starting delete reward type")
	rewardType, err := r.rep.GetRewardType(ctx, rewardTypeID)
	if err != nil {
		return err
	}

	err = CheckShopAdminAccess(ctx, logger, r.workerCsRep, deleterID, *rewardType.CoffeeShopID)
	if err != nil {
		return err
	}

	err = r.rep.DeleteReward(ctx, rewardTypeID)
	if err != nil {
		logger.Error("unexpected error when deleting reward type")
		return err
	}
	logger.Info("reward type deleted successfully")
	return nil
}

// GetRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) GetRewardType(ctx context.Context, rewardTypeID uuid.UUID) (*dto.RewardTypeResponse, error) {
	logger := r.logger.With("method", "GetRewardType", "reward type id", rewardTypeID.String())
	logger.Debug("starting get reward type")
	rewardType, err := r.rep.GetRewardType(ctx, rewardTypeID)
	if err != nil {
		return nil, err
	}

	return toRewardTypeResponse(rewardType), nil
}

// GetRewardsTypesFromCoffeeShop implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) GetRewardsTypesFromCoffeeShop(ctx context.Context, coffeeShopID uuid.UUID, page int, limit int) ([]dto.RewardTypeResponse, error) {
	logger := r.logger.With("method", "GetRewardsTypeFromCoffeeShopID", "coffee shop id", coffeeShopID.String())
	logger.Debug("starting get reward type from coffee shop id")
	rewardsTypes, err := r.rep.GetRewardsTypeByCoffeeShopID(ctx, coffeeShopID, limit*page, limit)
	if err != nil {
		return nil, err
	}

	return toRewardsTypesResponses(rewardsTypes), nil
}

// UpdateRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) UpdateRewardType(ctx context.Context, updaterID uuid.UUID, rewardTypeID uuid.UUID, request *dto.UpdateRewardTypeRequest) error {
	logger := r.logger.With("method", "UpdateRewardType", "updater id", updaterID.String())
	logger.Debug("starting update reward type")
	rewardType, err := r.rep.GetRewardType(ctx, rewardTypeID)
	if err != nil {
		return err
	}

	err = CheckShopAdminAccess(ctx, logger, r.workerCsRep, updaterID, *rewardType.CoffeeShopID)
	if err != nil {
		return err
	}

	if request.Description != nil {
		rewardType.Description = *request.Description
	}

	err = r.rep.UpdateReward(ctx, rewardType)
	if err != nil {
		return err
	}

	logger.Info("reward type updated successfully")
	return nil
}

func toRewardType(request *dto.CreateRewardTypeRequest) *models.RewardType {
	return &models.RewardType{
		CoffeeShopID: &request.CoffeeShopID,
		Description:  request.Description,
	}
}

func toRewardTypeResponse(rewardType *models.RewardType) *dto.RewardTypeResponse {
	return &dto.RewardTypeResponse{
		ID:           rewardType.ID,
		CoffeeShopID: *rewardType.CoffeeShopID,
		Description:  rewardType.Description,
	}
}

func toRewardsTypesResponses(rewardsTypes []models.RewardType) []dto.RewardTypeResponse {
	responses := make([]dto.RewardTypeResponse, len(rewardsTypes))
	for i := range rewardsTypes {
		responses[i] = dto.RewardTypeResponse{
			ID:           rewardsTypes[i].ID,
			CoffeeShopID: *rewardsTypes[i].CoffeeShopID,
			Description:  rewardsTypes[i].Description,
		}
	}

	return responses
}
