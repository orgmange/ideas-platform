package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type IdeaRepository interface {
	CreateIdea(*models.Idea) (*models.Idea, error)
	GetIdea(ideaID uuid.UUID) (*models.Idea, error)
	GetAllIdeasByShop(shopID uuid.UUID, limit, offset int, sort string) ([]models.Idea, error)
	GetAllIdeasByUser(userID uuid.UUID, limit, offset int, sort string) ([]models.Idea, error)
	UpdateIdea(*models.Idea) error
	DeleteIdea(IdeaID uuid.UUID) error
}

