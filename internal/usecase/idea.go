package usecase

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type IdeaUsecase interface {
	GetIdea(ctx context.Context, ideaID uuid.UUID) (*dto.IdeaResponse, error)
	GetAllIdeasByShop(ctx context.Context, shopID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error)
	GetAllIdeasByUser(ctx context.Context, userID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error)
	CreateIdea(ctx context.Context, userID uuid.UUID, req *dto.CreateIdeaRequest) (*dto.IdeaResponse, error)
	UpdateIdea(ctx context.Context, userID, ideaID uuid.UUID, req *dto.UpdateIdeaRequest) error
	DeleteIdea(ctx context.Context, userID, ideaID uuid.UUID) error
}
