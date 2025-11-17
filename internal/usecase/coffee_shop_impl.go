package usecase

import (
	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type CoffeeShopUsecaseImpl struct {
	rep repository.CoffeeShopRep
}

func NewCoffeeShopUsecase(rep repository.CoffeeShopRep) CoffeeShopUsecase {
	return &CoffeeShopUsecaseImpl{rep: rep}
}

func (u *CoffeeShopUsecaseImpl) CreateCoffeeShop(userID uuid.UUID, req *dto.CreateCoffeeShopRequest) (*dto.CoffeeShopResponse, error) {
	shop := toCoffeeShop(req)
	shop.CreatorID = userID
	createdShop, err := u.rep.CreateCoffeeShop(shop)
	if err != nil {
		return nil, err
	}

	return toCoffeeShopResponse(createdShop), nil
}

func (u *CoffeeShopUsecaseImpl) DeleteCoffeeShop(userID uuid.UUID, ID uuid.UUID) error {
	_, err := u.getIfCreator(userID, ID)
	if err != nil {
		return err
	}
	return u.rep.DeleteCoffeeShop(ID)
}

func (u *CoffeeShopUsecaseImpl) GetAllCoffeeShops(page int, limit int) ([]dto.CoffeeShopResponse, error) {
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}
	shops, err := u.rep.GetAllCoffeeShops(limit, limit*page)
	if err != nil {
		return nil, err
	}
	return toCoffeeShopResponses(shops), nil
}

func (u *CoffeeShopUsecaseImpl) GetCoffeeShop(ID uuid.UUID) (*dto.CoffeeShopResponse, error) {
	shop, err := u.rep.GetCoffeeShop(ID)
	if err != nil {
		return nil, err
	}

	return toCoffeeShopResponse(shop), nil
}

func (u *CoffeeShopUsecaseImpl) UpdateCoffeeShop(userID uuid.UUID, ID uuid.UUID, req *dto.UpdateCoffeeShopRequest) error {
	shop, err := u.getIfCreator(userID, ID)
	if err != nil {
		return err
	}

	if req.Name != "" {
		shop.Name = req.Name
	}
	if req.Address != "" {
		shop.Address = req.Address
	}
	if req.Contacts != nil {
		shop.Contacts = req.Contacts
	}
	if req.WelcomeMessage != nil {
		shop.WelcomeMessage = req.WelcomeMessage
	}
	if req.Rules != nil {
		shop.Rules = req.Rules
	}

	return u.rep.UpdateCoffeeShop(shop)
}

func toCoffeeShop(req *dto.CreateCoffeeShopRequest) *models.CoffeeShop {
	return &models.CoffeeShop{
		Name:           req.Name,
		Address:        req.Address,
		Contacts:       req.Contacts,
		WelcomeMessage: req.WelcomeMessage,
		Rules:          req.Rules,
	}
}

func toCoffeeShopResponse(shop *models.CoffeeShop) *dto.CoffeeShopResponse {
	return &dto.CoffeeShopResponse{
		ID:             shop.ID,
		Name:           shop.Name,
		Address:        shop.Address,
		Contacts:       shop.Contacts,
		WelcomeMessage: shop.WelcomeMessage,
		Rules:          shop.Rules,
	}
}

func toCoffeeShopResponses(shops []models.CoffeeShop) []dto.CoffeeShopResponse {
	res := make([]dto.CoffeeShopResponse, len(shops))
	for i := range shops {
		res[i] = *toCoffeeShopResponse(&shops[i])
	}

	return res
}

func (u *CoffeeShopUsecaseImpl) getIfCreator(userID uuid.UUID, shopID uuid.UUID) (*models.CoffeeShop, error) {
	shop, err := u.rep.GetCoffeeShop(shopID)
	if err != nil {
		return nil, err
	}
	if userID != shop.CreatorID {
		return nil, apperrors.NewAuthErr("access denied")
	}
	return shop, nil
}
