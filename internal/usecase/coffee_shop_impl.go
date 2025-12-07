package usecase

import (
	"context"
	"errors"
	"log/slog"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type CoffeeShopUsecaseImpl struct {
	rep         repository.CoffeeShopRep
	workerCsRep repository.WorkerCoffeeShopRepository
	adminRoleID uuid.UUID // Added adminRoleID
	logger      *slog.Logger
}

func NewCoffeeShopUsecase(rep repository.CoffeeShopRep,
	workerCsRep repository.WorkerCoffeeShopRepository,
	adminRoleID uuid.UUID,
	logger *slog.Logger,
) CoffeeShopUsecase {
	return &CoffeeShopUsecaseImpl{
		rep:         rep,
		workerCsRep: workerCsRep,
		adminRoleID: adminRoleID,
		logger:      logger,
	}
}

func (u *CoffeeShopUsecaseImpl) CreateCoffeeShop(ctx context.Context, userID uuid.UUID, req *dto.CreateCoffeeShopRequest) (*dto.CoffeeShopResponse, error) {
	logger := u.logger.With("method", "CreateCoffeeShop", "userID", userID.String())
	logger.Debug("starting create coffee shop")

	shop := toCoffeeShop(req)
	shop.CreatorID = userID
	createdShop, err := u.rep.CreateCoffeeShop(ctx, shop)
	if err != nil {
		logger.Error("failed to create coffee shop", "error", err.Error())
		return nil, err
	}
	_, err = u.workerCsRep.Create(ctx, &models.WorkerCoffeeShop{
		CoffeeShopID: &createdShop.ID,
		WorkerID:     &userID,
		RoleID:       &u.adminRoleID, // Assign AdminRoleID
	})
	if err != nil {
		return nil, err
	}
	logger.Info("coffee shop created successfully", "shopID", createdShop.ID.String())
	return toCoffeeShopResponse(createdShop), nil
}

func (u *CoffeeShopUsecaseImpl) DeleteCoffeeShop(ctx context.Context, userID uuid.UUID, ID uuid.UUID) error {
	logger := u.logger.With("method", "DeleteCoffeeShop", "userID", userID.String(), "shopID", ID.String())
	logger.Debug("starting delete coffee shop")

	_, err := u.getIfCreator(ctx, userID, ID)
	if err != nil {
		return err
	}
	err = u.rep.DeleteCoffeeShop(ctx, ID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("coffee shop to delete not found")
			return err
		}
		logger.Error("failed to delete coffee shop", "error", err.Error())
		return err
	}

	logger.Info("coffee shop deleted successfully")
	return nil
}

func (u *CoffeeShopUsecaseImpl) GetAllCoffeeShops(ctx context.Context, page int, limit int) ([]dto.CoffeeShopResponse, error) {
	logger := u.logger.With("method", "GetAllCoffeeShops", "page", page, "limit", limit)
	logger.Debug("starting get all coffee shops")

	if limit <= 0 || limit > 25 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}
	shops, err := u.rep.GetAllCoffeeShops(ctx, limit, limit*page)
	if err != nil {
		logger.Error("failed to get all coffee shops", "error", err.Error())
		return nil, err
	}

	logger.Info("coffee shops fetched successfully", "count", len(shops))
	return toCoffeeShopResponses(shops), nil
}

func (u *CoffeeShopUsecaseImpl) GetCoffeeShop(ctx context.Context, ID uuid.UUID) (*dto.CoffeeShopResponse, error) {
	logger := u.logger.With("method", "GetCoffeeShop", "shopID", ID.String())
	logger.Debug("starting get coffee shop")

	shop, err := u.rep.GetCoffeeShop(ctx, ID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("coffee shop not found")
			return nil, err
		}
		logger.Error("failed to get coffee shop", "error", err.Error())
		return nil, err
	}

	logger.Info("coffee shop fetched successfully")
	return toCoffeeShopResponse(shop), nil
}

func (u *CoffeeShopUsecaseImpl) UpdateCoffeeShop(ctx context.Context, userID uuid.UUID, ID uuid.UUID, req *dto.UpdateCoffeeShopRequest) error {
	logger := u.logger.With("method", "UpdateCoffeeShop", "userID", userID.String(), "shopID", ID.String())
	logger.Debug("starting update coffee shop")

	shop, err := u.getIfCreator(ctx, userID, ID)
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

	err = u.rep.UpdateCoffeeShop(ctx, shop)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("coffee shop to update not found")
			return err
		}
		logger.Error("failed to update coffee shop", "error", err.Error())
		return err
	}

	logger.Info("coffee shop updated successfully")
	return nil
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

func (u *CoffeeShopUsecaseImpl) getIfCreator(ctx context.Context, userID uuid.UUID, shopID uuid.UUID) (*models.CoffeeShop, error) {
	logger := u.logger.With("method", "getIfCreator", "userID", userID.String(), "shopID", shopID.String())
	logger.Debug("checking if user is creator of coffee shop")

	shop, err := u.rep.GetCoffeeShop(ctx, shopID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("coffee shop not found for creator check")
			return nil, err
		}
		logger.Error("failed to get coffee shop for creator check", "error", err.Error())
		return nil, err
	}
	if userID != shop.CreatorID {
		logger.Info("user is not creator of the coffee shop", "creatorID", shop.CreatorID)
		return nil, apperrors.NewErrUnauthorized("access denied")
	}
	logger.Debug("user is confirmed as creator")
	return shop, nil
}
