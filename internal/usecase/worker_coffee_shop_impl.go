package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type WorkerCoffeeShopUsecaseImpl struct {
	workerShopRepo repository.WorkerCoffeeShopRepository
	coffeeShopRepo repository.CoffeeShopRep
	userRepo       repository.UserRep
	logger         *slog.Logger
}

func NewWorkerCoffeeShopUsecase(
	workerShopRepo repository.WorkerCoffeeShopRepository,
	coffeeShopRepo repository.CoffeeShopRep,
	userRepo repository.UserRep,
	logger *slog.Logger,
) WorkerCoffeeShopUsecase {
	return &WorkerCoffeeShopUsecaseImpl{
		workerShopRepo: workerShopRepo,
		coffeeShopRepo: coffeeShopRepo,
		userRepo:       userRepo,
		logger:         logger,
	}
}

func (u *WorkerCoffeeShopUsecaseImpl) AddWorker(ctx context.Context, actorID uuid.UUID, req *dto.AddWorkerToShopRequest) (*dto.WorkerCoffeeShopResponse, error) {
	logger := u.logger.With("method", "AddWorker", "actorID", actorID, "workerID", req.WorkerID, "shopID", req.CoffeeShopID)
	logger.Debug("starting to add worker to shop")

	if err := u.checkShopAdminAccess(ctx, actorID, req.CoffeeShopID); err != nil {
		return nil, err
	}

	exists, err := u.userRepo.IsUserExist(ctx, req.WorkerID)
	if err != nil {
		logger.Error("failed to check user existence", "error", err)
		return nil, err
	}
	if !exists {
		logger.Info("user to add as worker does not exist")
		return nil, apperrors.NewErrNotFound("user", req.WorkerID.String())
	}

	_, err = u.workerShopRepo.GetByUserIDAndShopID(ctx, req.WorkerID, req.CoffeeShopID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if !errors.As(err, &errNotFound) {
			logger.Error("failed to check if user is already a worker", "error", err)
			return nil, err
		}
	} else {
		logger.Info("user is already a worker in this shop")
		return nil, apperrors.NewErrConflict(fmt.Sprintf("user %s is already a worker in shop %s", req.WorkerID, req.CoffeeShopID))
	}

	relation := &models.WorkerCoffeeShop{
		WorkerID:     &req.WorkerID,
		CoffeeShopID: &req.CoffeeShopID,
	}

	createdRelation, err := u.workerShopRepo.Create(ctx, relation)
	if err != nil {
		logger.Error("failed to create worker-shop relation", "error", err)
		return nil, err
	}

	logger.Info("worker added successfully")
	return toWorkerCoffeeShopResponse(createdRelation), nil
}

func (u *WorkerCoffeeShopUsecaseImpl) RemoveWorker(ctx context.Context, actorID, workerShopRelationID uuid.UUID) error {
	logger := u.logger.With("method", "RemoveWorker", "actorID", actorID, "relationID", workerShopRelationID)
	logger.Debug("starting to remove worker from shop")

	relation, err := u.workerShopRepo.GetByID(ctx, workerShopRelationID)
	if err != nil {
		return err // Error already logged and classified by repository
	}

	if err := u.checkShopAdminAccess(ctx, actorID, *relation.CoffeeShopID); err != nil {
		return err
	}

	if err := u.workerShopRepo.Delete(ctx, workerShopRelationID); err != nil {
		logger.Error("failed to delete worker-shop relation", "error", err)
		return err
	}

	logger.Info("worker removed successfully")
	return nil
}

func (u *WorkerCoffeeShopUsecaseImpl) ListWorkers(ctx context.Context, actorID, shopID uuid.UUID, page, limit int) ([]dto.UserResponse, error) {
	logger := u.logger.With("method", "ListWorkers", "actorID", actorID, "shopID", shopID, "page", page, "limit", limit)
	logger.Debug("starting to list workers in shop")

	if err := u.checkShopAdminAccess(ctx, actorID, shopID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 50 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}

	relations, err := u.workerShopRepo.ListByCoffeeShopID(ctx, shopID, limit, page*limit)
	if err != nil {
		logger.Error("failed to list workers by coffee shop id", "error", err)
		return nil, err
	}

	logger.Info("workers listed successfully", "count", len(relations))
	return toUserResponsesFromRelations(relations), nil
}

func (u *WorkerCoffeeShopUsecaseImpl) ListShopsForWorker(ctx context.Context, actorID, workerID uuid.UUID, page, limit int) ([]dto.CoffeeShopResponse, error) {
	logger := u.logger.With("method", "ListShopsForWorker", "actorID", actorID, "workerID", workerID, "page", page, "limit", limit)
	logger.Debug("starting to list shops for worker")

	if actorID != workerID {
		logger.Warn("access denied: user trying to access other user's data", "actorID", actorID, "targetWorkerID", workerID)
		return nil, apperrors.NewErrAccessDenied("you can only view your own coffee shops")
	}

	if limit <= 0 || limit > 50 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}

	relations, err := u.workerShopRepo.ListByWorkerID(ctx, workerID, limit, page*limit)
	if err != nil {
		logger.Error("failed to list shops by worker id", "error", err)
		return nil, err
	}

	logger.Info("shops for worker listed successfully", "count", len(relations))
	return toCoffeeShopResponsesFromRelations(relations), nil
}

// checkShopAdminAccess verifies if a user is either the creator of the shop,
// or a worker in the shop with the 'admin' role.
func (u *WorkerCoffeeShopUsecaseImpl) checkShopAdminAccess(ctx context.Context, actorID, shopID uuid.UUID) error {
	logger := u.logger.With("method", "checkShopAdminAccess", "actorID", actorID, "shopID", shopID)

	worker, err := u.workerShopRepo.GetByUserIDAndShopID(ctx, actorID, shopID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			return apperrors.NewErrAccessDenied("user is not worker for this coffee shop")
		}
		return err
	}
	if worker.Role.Name == "admin" {
		logger.Debug("access granted: user is a worker with admin role")
		return nil
	}

	logger.Warn("access denied: user is not shop creator or admin worker")
	return apperrors.NewErrAccessDenied("user is not an admin for this coffee shop")
}

// --- DTO Mappers ---

func toWorkerCoffeeShopResponse(r *models.WorkerCoffeeShop) *dto.WorkerCoffeeShopResponse {
	return &dto.WorkerCoffeeShopResponse{
		ID: r.ID,
		Worker: dto.UserResponse{
			ID:    r.Worker.ID,
			Name:  r.Worker.Name,
			Phone: r.Worker.Phone,
		},
		CoffeeShop: dto.CoffeeShopResponse{
			ID:             r.CoffeeShop.ID,
			Name:           r.CoffeeShop.Name,
			Address:        r.CoffeeShop.Address,
			Contacts:       r.CoffeeShop.Contacts,
			WelcomeMessage: r.CoffeeShop.WelcomeMessage,
			Rules:          r.CoffeeShop.Rules,
		},
	}
}

func toUserResponsesFromRelations(relations []models.WorkerCoffeeShop) []dto.UserResponse {
	users := make([]dto.UserResponse, len(relations))
	for i, r := range relations {
		users[i] = dto.UserResponse{
			ID:    r.Worker.ID,
			Name:  r.Worker.Name,
			Phone: r.Worker.Phone,
		}
	}
	return users
}

func toCoffeeShopResponsesFromRelations(relations []models.WorkerCoffeeShop) []dto.CoffeeShopResponse {
	shops := make([]dto.CoffeeShopResponse, len(relations))
	for i, r := range relations {
		shops[i] = dto.CoffeeShopResponse{
			ID:             r.CoffeeShop.ID,
			Name:           r.CoffeeShop.Name,
			Address:        r.CoffeeShop.Address,
			Contacts:       r.CoffeeShop.Contacts,
			WelcomeMessage: r.CoffeeShop.WelcomeMessage,
			Rules:          r.CoffeeShop.Rules,
		}
	}
	return shops
}
