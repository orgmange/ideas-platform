package repository

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type WorkerCoffeeShopRepository interface {
	Create(ctx context.Context, workerShop *models.WorkerCoffeeShop) (*models.WorkerCoffeeShop, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.WorkerCoffeeShop, error)
	ListByCoffeeShopID(ctx context.Context, coffeeShopID uuid.UUID, limit, offset int) ([]models.WorkerCoffeeShop, error)
	ListByWorkerID(ctx context.Context, workerID uuid.UUID, limit, offset int) ([]models.WorkerCoffeeShop, error)
	Update(ctx context.Context, workerShop *models.WorkerCoffeeShop) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserIDAndShopID(ctx context.Context, userID, shopID uuid.UUID) (*models.WorkerCoffeeShop, error)
	IsAdminInAnyShop(ctx context.Context, userID uuid.UUID) (bool, error)
}
