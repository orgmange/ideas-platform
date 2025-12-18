package usecase

import (
	"context"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type IdeaStatusUsecase interface {
	Create(ctx context.Context, status dto.CreateIdeaStatusRequest) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, status dto.UpdateIdeaStatusRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (dto.IdeaStatusResponse, error)
	GetAll(ctx context.Context) ([]dto.IdeaStatusResponse, error)
}
