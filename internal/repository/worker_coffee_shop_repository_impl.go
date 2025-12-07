package repository

import (
	"context"
	"errors"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkerCoffeeShopRepositoryImpl struct {
	db *gorm.DB
}

func NewWorkerCoffeeShopRepository(db *gorm.DB) WorkerCoffeeShopRepository {
	return &WorkerCoffeeShopRepositoryImpl{db: db}
}

// Create creates a new worker-coffeeshop relationship
func (r *WorkerCoffeeShopRepositoryImpl) Create(ctx context.Context, workerShop *models.WorkerCoffeeShop) (*models.WorkerCoffeeShop, error) {
	if err := r.db.WithContext(ctx).Create(workerShop).Error; err != nil {
		return nil, err
	}
	// Preload associated data for the returned object
	if err := r.db.WithContext(ctx).Preload("Worker").Preload("CoffeeShop").First(workerShop).Error; err != nil {
		return nil, err
	}
	return workerShop, nil
}

// GetByID retrieves a worker-coffeeshop relationship by its ID
func (r *WorkerCoffeeShopRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkerCoffeeShop, error) {
	var workerShop models.WorkerCoffeeShop
	err := r.db.WithContext(ctx).Preload("Worker").Preload("CoffeeShop").
		Where("is_deleted = ?", false).First(&workerShop, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewErrNotFound("worker_coffee_shop", id.String())
		}
		return nil, err
	}
	return &workerShop, nil
}

// ListByCoffeeShopID lists all workers for a given coffee shop
func (r *WorkerCoffeeShopRepositoryImpl) ListByCoffeeShopID(ctx context.Context, coffeeShopID uuid.UUID, limit, offset int) ([]models.WorkerCoffeeShop, error) {
	var workers []models.WorkerCoffeeShop
	err := r.db.WithContext(ctx).Preload("Worker").
		Where("coffee_shop_id = ? AND is_deleted = ?", coffeeShopID, false).
		Limit(limit).Offset(offset).
		Find(&workers).Error
	return workers, err
}

// ListByWorkerID lists all coffee shops for a given worker
func (r *WorkerCoffeeShopRepositoryImpl) ListByWorkerID(ctx context.Context, workerID uuid.UUID, limit, offset int) ([]models.WorkerCoffeeShop, error) {
	var coffeeShops []models.WorkerCoffeeShop
	err := r.db.WithContext(ctx).Preload("CoffeeShop").
		Where("worker_id = ? AND is_deleted = ?", workerID, false).
		Limit(limit).Offset(offset).
		Find(&coffeeShops).Error
	return coffeeShops, err
}

// Update updates a worker-coffeeshop relationship
func (r *WorkerCoffeeShopRepositoryImpl) Update(ctx context.Context, workerShop *models.WorkerCoffeeShop) error {
	return r.db.WithContext(ctx).Save(workerShop).Error
}

// Delete soft-deletes a worker-coffeeshop relationship
func (r *WorkerCoffeeShopRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&models.WorkerCoffeeShop{}).Where("id = ?", id).
		Update("is_deleted", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("worker_coffee_shop", id.String())
	}
	return nil
}

func (r *WorkerCoffeeShopRepositoryImpl) GetByUserIDAndShopID(ctx context.Context, userID, shopID uuid.UUID) (*models.WorkerCoffeeShop, error) {
	var worker *models.WorkerCoffeeShop
	err := r.db.WithContext(ctx).Preload("Worker").Preload("CoffeeShop").Preload("Role").
		Where("worker_id = ? AND coffee_shop_id = ? AND is_deleted = ?", userID, shopID, false).
		First(&worker).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewErrNotFound("worker coffee shop", "user ID: "+userID.String()+" coffee shop ID: "+shopID.String())
		}
		return nil, err
	}
	return worker, nil
}

func (r *WorkerCoffeeShopRepositoryImpl) IsAdminInAnyShop(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.WorkerCoffeeShop{}).
		Joins("JOIN role ON role.id = worker_coffee_shop.role_id").
		Where("worker_coffee_shop.worker_id = ? AND role.name = ? AND worker_coffee_shop.is_deleted = ?", userID, "admin", false).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
