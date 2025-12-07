package usecase

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type CoffeeShopUsecase interface {
	GetCoffeeShop(ctx context.Context, id uuid.UUID) (*dto.CoffeeShopResponse, error)
	GetAllCoffeeShops(ctx context.Context, page, pageSize int) ([]dto.CoffeeShopResponse, error)
	CreateCoffeeShop(ctx context.Context, userID uuid.UUID, req *dto.CreateCoffeeShopRequest) (*dto.CoffeeShopResponse, error)
	UpdateCoffeeShop(ctx context.Context, userID uuid.UUID, ID uuid.UUID, req *dto.UpdateCoffeeShopRequest) error
	DeleteCoffeeShop(ctx context.Context, userID uuid.UUID, ID uuid.UUID) error
}
