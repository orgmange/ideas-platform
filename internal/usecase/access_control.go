package usecase

import (
	"context"
	"errors"
	"log/slog"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

// CheckShopAdminAccess verifies if a user is a worker in the shop with the 'admin' role.
func CheckShopAdminAccess(ctx context.Context, logger *slog.Logger, workerShopRepo repository.WorkerCoffeeShopRepository, actorID, shopID uuid.UUID) error {
	l := logger.With("method", "CheckShopAdminAccess", "actorID", actorID, "shopID", shopID)

	worker, err := workerShopRepo.GetByUserIDAndShopID(ctx, actorID, shopID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			return apperrors.NewErrAccessDenied("user is not worker for this coffee shop")
		}
		return err
	}
	if worker.Role.Name == "admin" {
		l.Debug("access granted: user is a worker with admin role")
		return nil
	}

	l.Warn("access denied: user is not shop creator or admin worker")
	return apperrors.NewErrAccessDenied("user is not an admin for this coffee shop")
}

// CheckAnyShopAdminAccess verifies if a user is an admin in at least one coffee shop.
func CheckAnyShopAdminAccess(ctx context.Context, logger *slog.Logger, workerShopRepo repository.WorkerCoffeeShopRepository, actorID uuid.UUID) error {
	l := logger.With("method", "CheckAnyShopAdminAccess", "actorID", actorID)

	isAdmin, err := workerShopRepo.IsAdminInAnyShop(ctx, actorID)
	if err != nil {
		return err
	}

	if isAdmin {
		l.Debug("access granted: user is an admin in at least one shop")
		return nil
	}

	l.Warn("access denied: user is not an admin in any coffee shop")
	return apperrors.NewErrAccessDenied("user is not an admin in any coffee shop")
}
