package usecase

import (
	"log/slog"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type RewardTypeUsecaseImpl struct {
	rep    repository.RewardTypeRepository
	csRep  repository.CoffeeShopRep
	logger *slog.Logger
}

func NewRewardTypeUsecase(rep repository.RewardTypeRepository,
	csRep repository.CoffeeShopRep,
	logger *slog.Logger,
) RewardTypeUsecase {
	return &RewardTypeUsecaseImpl{
		rep:    rep,
		csRep:  csRep,
		logger: logger,
	}
}

// CreateRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) CreateRewardType(creatorID uuid.UUID, creatorRole string, request *dto.CreateRewardTypeRequest) (*dto.RewardTypeResponse, error) {
	logger := r.logger.With("method", "CreateRewardType", "creator id", creatorID.String(), "creator role", creatorRole, "coffee shop id", request.CoffeeShopID.String())
	logger.Debug("starting create reward type")

	coffeeShop, err := r.csRep.GetCoffeeShop(request.CoffeeShopID)
	if err != nil {
		return nil, err
	}

	err = r.checkAccessToRewardType(creatorID, coffeeShop, creatorRole)
	if err != nil {
		return nil, err
	}

	rewardType := toRewardType(request)
	savedRewardType, err := r.rep.CreateReward(rewardType)
	if err != nil {
		logger.Error("unexpected error when creating reward type", "error", err.Error())
		return nil, err
	}

	logger.Error("reward type created successfully", "reward type id", savedRewardType.ID.String())
	return toRewardTypeResponse(savedRewardType), nil
}

// DeleteRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) DeleteRewardType(deleterID uuid.UUID, deleterRole string, rewardTypeID uuid.UUID) error {
	logger := r.logger.With("method", "DeleteRewardType", "deleter id", deleterID.String(), "deleter role", deleterRole)
	logger.Debug("starting delete reward type")
	rewardType, err := r.rep.GetRewardType(rewardTypeID)
	if err != nil {
		return err
	}

	err = r.checkAccessToRewardType(deleterID, rewardType.CoffeeShop, deleterRole)
	if err != nil {
		return err
	}

	err = r.rep.DeleteReward(rewardTypeID)
	if err != nil {
		logger.Error("unexpected error when deleting reward type")
		return err
	}
	logger.Info("reward type deleted successfully")
	return nil
}

// GetRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) GetRewardType(rewardTypeID uuid.UUID) (*dto.RewardTypeResponse, error) {
	logger := r.logger.With("method", "GetRewardType", "reward type id", rewardTypeID.String())
	logger.Debug("starting get reward type")
	rewardType, err := r.rep.GetRewardType(rewardTypeID)
	if err != nil {
		return nil, err
	}

	return toRewardTypeResponse(rewardType), nil
}

// GetRewardsTypesFromCoffeeShop implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) GetRewardsTypesFromCoffeeShop(coffeeShopID uuid.UUID, page int, limit int) ([]dto.RewardTypeResponse, error) {
	logger := r.logger.With("method", "GetRewardsTypeFromCoffeeShopID", "coffee shop id", coffeeShopID.String())
	logger.Debug("starting get reward type from coffee shop id")
	rewardsTypes, err := r.rep.GetRewardsTypeByCoffeeShopID(coffeeShopID, limit*page, limit)
	if err != nil {
		return nil, err
	}

	return toRewardsTypesResponses(rewardsTypes), nil
}

// UpdateRewardType implements RewardTypeUsecase.
func (r *RewardTypeUsecaseImpl) UpdateRewardType(updaterID uuid.UUID, updaterRole string, rewardTypeID uuid.UUID, request *dto.UpdateRewardTypeRequest) error {
	logger := r.logger.With("method", "UpdateRewardType", "updater id", updaterID.String(), "updater role", updaterRole)
	logger.Debug("starting update reward type")
	rewardType, err := r.rep.GetRewardType(rewardTypeID)
	if err != nil {
		return err
	}

	err = r.checkAccessToRewardType(updaterID, rewardType.CoffeeShop, updaterRole)
	if err != nil {
		return err
	}

	if request.Description != nil {
		rewardType.Description = *request.Description
	}

	err = r.rep.UpdateReward(rewardType)
	if err != nil {
		return err
	}

	logger.Info("reward type updated successfully")
	return nil
}

func isWorker(userID uuid.UUID, coffeeShop *models.CoffeeShop) (bool, error) {
	// TODO: реализовать метод проверки является ли этот пользователь работником этого кофешопа
	return true, nil
}

func (r *RewardTypeUsecaseImpl) checkAccessToRewardType(userID uuid.UUID, coffeeShop *models.CoffeeShop, userRole string) error {
	logger := r.logger.With("method", "checkAccessToRewardType", "user id", userID.String(), "user role", userRole, "coffee shop id", coffeeShop.ID.String())
	logger.Debug("starting check access to reward type")

	if userRole != "admin" {
		logger.Info("thst user dont have permissons for reward type")
		return apperrors.NewErrAccessDenied("forbidden")
	}

	isWorker, err := isWorker(userID, coffeeShop)
	if err != nil {
		return err
	}
	if !isWorker {
		logger.Info("that user is not worker for this coffee shop")
		return apperrors.NewErrAccessDenied("that user is not worker for this coffee shop")
	}

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
