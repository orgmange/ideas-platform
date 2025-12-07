package repository

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CoffeeShopRepImpl struct {
	db *gorm.DB
}

func NewCoffeeShopRepository(db *gorm.DB) CoffeeShopRep {
	return &CoffeeShopRepImpl{db: db}
}

func (r *CoffeeShopRepImpl) CreateCoffeeShop(ctx context.Context, shop *models.CoffeeShop) (*models.CoffeeShop, error) {
	if err := r.db.WithContext(ctx).Create(shop).Error; err != nil {
		return nil, err
	}
	return shop, nil
}

func (r *CoffeeShopRepImpl) UpdateCoffeeShop(ctx context.Context, shop *models.CoffeeShop) error {
	result := r.db.WithContext(ctx).Save(shop)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("coffee_shop", shop.ID.String())
	}
	return nil
}

func (r *CoffeeShopRepImpl) DeleteCoffeeShop(ctx context.Context, ID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.CoffeeShop{}, ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("coffee_shop", ID.String())
	}
	return nil
}

func (r *CoffeeShopRepImpl) GetCoffeeShop(ctx context.Context, ID uuid.UUID) (*models.CoffeeShop, error) {
	var shop models.CoffeeShop
	if err := r.db.WithContext(ctx).First(&shop, "id = ?", ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewErrNotFound("coffee_shop", ID.String())
		}
		return nil, err
	}
	return &shop, nil
}

func (r *CoffeeShopRepImpl) GetAllCoffeeShops(ctx context.Context, limit, offset int) ([]models.CoffeeShop, error) {
	var shops []models.CoffeeShop
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&shops).Error; err != nil {
		return nil, err
	}
	return shops, nil
}

func (r *CoffeeShopRepImpl) IsCoffeeShopExist(ctx context.Context, ID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CoffeeShop{}).Where("id = ?", ID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check coffee shop existence: %w", err)
	}
	return count > 0, nil
}

// IsWorker implements CoffeeShopRep.
func (r *CoffeeShopRepImpl) IsWorker(ctx context.Context, userID uuid.UUID, shopID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WorkerCoffeeShop{}).Where("worker_id = ? and coffee_shop_id = ?", userID, shopID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
