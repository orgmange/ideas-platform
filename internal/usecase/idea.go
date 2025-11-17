package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type IdeaUsecase interface {
	GetIdea(ideaID uuid.UUID) (*dto.IdeaResponse, error)
	GetAllIdeasByShop(shopID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error)
	GetAllIdeasByUser(userID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error)
	CreateIdea(userID uuid.UUID, req *dto.CreateIdeaRequest) (*dto.IdeaResponse, error)
	UpdateIdea(userID, ideaID uuid.UUID, req *dto.UpdateIdeaRequest) error
	DeleteIdea(userID, ideaID uuid.UUID) error
}
