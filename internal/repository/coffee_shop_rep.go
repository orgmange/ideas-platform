package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type CoffeeShopRep interface {
	CreateCoffeeShop(*models.CoffeeShop) (*models.CoffeeShop, error)
	UpdateCoffeeShop(*models.CoffeeShop) error
	DeleteCoffeeShop(ID uuid.UUID) error
	GetCoffeeShop(ID uuid.UUID) (*models.CoffeeShop, error)
	GetAllCoffeeShops(limit, offset int) ([]models.CoffeeShop, error)
	IsCoffeeShopExist(ID uuid.UUID) (bool, error)
}
