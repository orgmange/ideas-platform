package usecase

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type WorkerCoffeeShopUsecase interface {
	// AddWorker adds a user as a worker to a coffee shop.
	// Only the coffee shop creator or an admin of that shop can perform this action.
	AddWorker(ctx context.Context, actorID uuid.UUID, req *dto.AddWorkerToShopRequest) (*dto.WorkerCoffeeShopResponse, error)

	// RemoveWorker removes a user from a coffee shop's workers.
	// The ID is the ID of the worker_coffee_shop relation.
	// Only the coffee shop creator or an admin of that shop can perform this action.
	RemoveWorker(ctx context.Context, actorID, workerShopRelationID uuid.UUID) error

	// ListWorkers retrieves a paginated list of workers for a specific coffee shop.
	// Requires admin access to the coffee shop.
	ListWorkers(ctx context.Context, actorID, shopID uuid.UUID, page, limit int) ([]dto.UserResponse, error)

	// ListShopsForWorker retrieves a paginated list of coffee shops a user works for.
	// This action is public.
	ListShopsForWorker(ctx context.Context, actorID, workerID uuid.UUID, page, limit int) ([]dto.CoffeeShopResponse, error)
}
