package repository

import (
	"context"

	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type IdeaStatusRepository interface {
	Create(ctx context.Context, status *models.IdeaStatus) (uuid.UUID, error)
	Update(ctx context.Context, status *models.IdeaStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (models.IdeaStatus, error)
	GetAll(ctx context.Context) ([]models.IdeaStatus, error)
	GetByTitle(ctx context.Context, title string) (models.IdeaStatus, error)
}
