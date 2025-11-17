package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type CoffeeShopUsecase interface {
	GetCoffeeShop(id uuid.UUID) (*dto.CoffeeShopResponse, error)
	GetAllCoffeeShops(page, pageSize int) ([]dto.CoffeeShopResponse, error)
	CreateCoffeeShop(userID uuid.UUID, req *dto.CreateCoffeeShopRequest) (*dto.CoffeeShopResponse, error)
	UpdateCoffeeShop(userID uuid.UUID, ID uuid.UUID, req *dto.UpdateCoffeeShopRequest) error
	DeleteCoffeeShop(userID uuid.UUID, ID uuid.UUID) error
}
