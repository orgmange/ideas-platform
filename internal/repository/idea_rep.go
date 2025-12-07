package repository

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type IdeaRepository interface {
	CreateIdea(ctx context.Context, idea *models.Idea) (*models.Idea, error)
	GetIdea(ctx context.Context, ideaID uuid.UUID) (*models.Idea, error)
	GetAllIdeasByShop(ctx context.Context, shopID uuid.UUID, limit, offset int, sort string) ([]models.Idea, error)
	GetAllIdeasByUser(ctx context.Context, userID uuid.UUID, limit, offset int, sort string) ([]models.Idea, error)
	UpdateIdea(ctx context.Context, idea *models.Idea) error
	DeleteIdea(ctx context.Context, IdeaID uuid.UUID) error
}
