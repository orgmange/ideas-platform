package repository

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type CoffeeShopRep interface {
	CreateCoffeeShop(ctx context.Context, shop *models.CoffeeShop) (*models.CoffeeShop, error)
	UpdateCoffeeShop(ctx context.Context, shop *models.CoffeeShop) error
	DeleteCoffeeShop(ctx context.Context, ID uuid.UUID) error
	GetCoffeeShop(ctx context.Context, ID uuid.UUID) (*models.CoffeeShop, error)
	GetAllCoffeeShops(ctx context.Context, limit, offset int) ([]models.CoffeeShop, error)
	IsCoffeeShopExist(ctx context.Context, ID uuid.UUID) (bool, error)
	IsWorker(ctx context.Context, userID, shopID uuid.UUID) (bool, error)
}
