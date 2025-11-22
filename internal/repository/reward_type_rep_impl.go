package repository

import (
	"errors"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardTypeRepositoryImpl struct {
	db *gorm.DB
}

func NewRewardTypeRepository(db *gorm.DB) RewardTypeRepository {
	return RewardTypeRepositoryImpl{
		db: db,
	}
}

func (r RewardTypeRepositoryImpl) CreateReward(rewardType *models.RewardType) (*models.RewardType, error) {
	err := r.db.Create(rewardType).Error
	if err != nil {
		return nil, err
	}

	return rewardType, nil
}

func (r RewardTypeRepositoryImpl) DeleteReward(rewardTypeID uuid.UUID) error {
	res := r.db.Delete(&models.RewardType{}, rewardTypeID)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return apperrors.NewErrNotFound("rewardType", rewardTypeID.String())
	}

	return nil
}

func (r RewardTypeRepositoryImpl) GetRewardType(rewardTypeID uuid.UUID) (*models.RewardType, error) {
	var rewardType models.RewardType
	err := r.db.Take(&rewardType, rewardTypeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewErrNotFound("rewardType", rewardTypeID.String())
		}
		return nil, err
	}

	return &rewardType, nil
}

func (r RewardTypeRepositoryImpl) GetRewardsTypeByCoffeeShopID(CoffeeShopID uuid.UUID, offset int, limit int) ([]models.RewardType, error) {
	var rewardTypes []models.RewardType
	err := r.db.Offset(offset).Limit(limit).Where("coffee_shop_id = ?", CoffeeShopID).Find(&rewardTypes).Error
	if err != nil {
		return rewardTypes, err
	}

	return rewardTypes, nil
}

func (r RewardTypeRepositoryImpl) UpdateReward(rewardType *models.RewardType) error {
	return r.db.Save(rewardType).Error
}
